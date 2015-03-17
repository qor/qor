package worker

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/qor/qor/utils"
)

var (
	WorkerDataPath = "worker_data"

	DefaultServerHost = func() string {
		ip, err := CurrentServerIP("") // current host ip, eth0 for linux and en0 for darwin
		if err != nil {
			fmt.Println("failed to retrieve current host ip")
		}
		return ip
	}()
	DefaultServerUser    = "app"
	DefaultServerSSHPort = "22"
)

func init() {
	if host := os.Getenv("QorJobHost"); host != "" {
		DefaultServerHost = host
	}
	if user := os.Getenv("QorJobUser"); user != "" {
		DefaultServerUser = user
	}
	if port := os.Getenv("QorJobPort"); port != "" {
		DefaultServerSSHPort = port
	}
}

// type JobStatus string

const (
	// Job statuses
	StatusToRun   = "Pending"
	StatusRunning = "Running"
	StatusFailed  = "Failed"
	StatusKilled  = "Killed"
	StatusDone    = "Done"
)

type QorJob struct {
	ID         uint64
	QueueJobId string

	// unit: minute
	// 0 to run job only once
	Interval uint64
	// zero time value to execute job immediately
	StartAt time.Time

	Stopped bool

	Cli        string
	WorkerName string
	JobName    string
	Status     string
	PID        int // TODO: change it into uint

	By string

	RunCounter     uint64
	FailCounter    uint64
	SuccessCounter uint64
	KillCounter    uint64

	ServerHost    string
	ServerUser    string
	ServerSSHPort string

	ExtraValue *ExtraInput `sql:"type:text;"` // Mysql: 64KB
	ExtraFile  *ExtraInput `sql:"type:text;"` // Mysql: 64KB

	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt time.Time

	Progress int

	log *os.File
}

type ExtraInput map[string]string

func (ei *ExtraInput) Scan(src interface{}) error {
	if ei == nil {
		ei = &ExtraInput{}
	}

	return json.Unmarshal(src.([]byte), ei)
}

func (ei *ExtraInput) Value() (driver.Value, error) {
	if ei == nil {
		return []byte{}, nil
	}

	return json.Marshal(ei)
}

func (ei ExtraInput) Open(name string) (f *os.File, err error) {
	return os.Open(filepath.Join(WorkerDataPath, name))
}

func (qj *QorJob) GetWorker() *Worker {
	if w, ok := workers[qj.WorkerName]; ok {
		return w
	}

	return nil
}

func (qj *QorJob) GetJob() *Job {
	if w, ok := workers[qj.WorkerName]; ok {
		return w.Jobs[qj.JobName]
	}

	// if j == nil {
	// 	err = fmt.Errorf("unknown job: %s:%s\n", qj.WorkerName, qj.JobName)
	// }

	return nil
}

func RunJob(jobId uint64) {
	job := &QorJob{}
	if err := jobDB.Find(job, jobId).Error; err != nil {
		fmt.Printf("job (%d) do not existed\n", jobId)
	} else {
		if job.Stopped {
			fmt.Println("can't run stopped job")
		} else {
			job.Run()
		}
	}
}

func (j *QorJob) Run() (err error) {
	parts := strings.Split(j.Cli, " ")
	name := parts[0]
	args := []string{"-job-id", strconv.FormatUint(j.ID, 10)}
	if len(parts) > 1 {
		args = append(args, parts[1:]...)
	}

	err = exec.Command(name, args...).Start()
	return
}

func (j *QorJob) UpdateStatus(status string) (err error) {
	old := j.Status
	j.Status = status
	changer := map[string]interface{}{"status": j.Status}
	switch status {
	case StatusRunning:
		j.RunCounter++
		changer["run_counter"] = j.RunCounter
	case StatusFailed:
		j.FailCounter++
		changer["fail_counter"] = j.FailCounter
	case StatusDone:
		j.SuccessCounter++
		changer["success_counter"] = j.SuccessCounter
	case StatusKilled:
		j.KillCounter++
		changer["kill_counter"] = j.KillCounter
	}

	if err = jobDB.Model(&QorJob{}).Where("id = ?", j.ID).UpdateColumns(changer).Error; err != nil {
		if logger, er := j.GetLogger(); er == nil {
			fmt.Fprintf(logger, "can't update status from %s to %s: %s\n", old, j.Status, err)
		}

		return
	}

	return
}

func (j *QorJob) GetLogger() (f *os.File, err error) {
	if j.log == nil {
		j.log, err = os.OpenFile(j.LogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	f = j.log

	return
}

func (qj *QorJob) LogPath() string {
	return fmt.Sprintf("%s/%s-%s-%d.log", WorkerDataPath, qj.WorkerName, qj.JobName, qj.ID)
}

func (j *QorJob) GetLog() (l string) {
	log, err := os.Open(j.LogPath())
	if err != nil {
		if !os.IsNotExist(err) {
			return "failed to open log: " + err.Error() + "."
		}
		return "log is empty."
	}
	lbytes, err := ioutil.ReadAll(log)
	if err != nil {
		return "failed to read log: " + err.Error() + "."
	}
	l = string(lbytes)
	return
}

// TODO: dequeue job will override value?
func (j *QorJob) SaveRunStatus() (err error) {
	j.ServerHost = DefaultServerHost
	j.ServerUser = DefaultServerUser
	j.ServerSSHPort = DefaultServerSSHPort
	j.PID = os.Getpid()
	changer := QorJob{
		ServerHost:    j.ServerHost,
		ServerUser:    j.ServerUser,
		ServerSSHPort: j.ServerSSHPort,
		PID:           j.PID,
		Status:        j.Status,
	}
	if err = jobDB.Model(j).UpdateColumns(changer).Error; err != nil {
		logger, erro := j.GetLogger()
		if erro == nil {
			fmt.Fprintf(logger, "can't save pid for job %d\n", j.ID)
		}

		return
	}

	return
}

func (q *QorJob) URL() string {
	w := q.GetWorker()
	if w == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%d", w.admin.GetRouter().Prefix, utils.ToParamString(w.Name), q.ID)
}

func CurrentServerIP(name string) (string, error) {
	if name == "" {
		name = "eth0"
		if runtime.GOOS == "darwin" {
			name = "en0"
		}
	}
	intfs, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	as, err := intfs.Addrs()
	if err != nil {
		return "", err
	}
	var ip net.IP
	for _, a := range as {
		ip, _, err = net.ParseCIDR(a.String())
		if err != nil {
			return "", err
		}
		if ip.To4() != nil {
			break
		}
	}
	return ip.String(), nil
}
