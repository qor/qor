package worker

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/qor/qor/utils"
)

var WorkerDataPath = "worker_data"

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
	Id         uint64
	QueueJobId string

	// unit: minute
	// 0 to run job only once
	Interval uint64

	// zero time value to execute job immediately
	StartAt time.Time

	Cli        string
	WorkerName string
	JobName    string
	Status     string
	PID        int

	By string

	RunCounter     uint64
	FailCounter    uint64
	SuccessCounter uint64
	KillCounter    uint64

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
		job.Run()
	}
}

func (j *QorJob) Run() (err error) {
	parts := strings.Split(j.Cli, " ")
	name := parts[0]
	args := []string{"-job-id", strconv.FormatUint(j.Id, 10)}
	if len(parts) > 1 {
		args = append(args, parts[1:]...)
	}

	err = exec.Command(name, args...).Start()
	return
}

func (j *QorJob) UpdateStatus(status string) (err error) {
	old := j.Status
	j.Status = status
	switch status {
	case StatusRunning:
		j.RunCounter++
	case StatusFailed:
		j.FailCounter++
	case StatusDone:
		j.SuccessCounter++
	case StatusKilled:
		j.KillCounter++
	}

	if err = jobDB.Save(j).Error; err != nil {
		logger, erro := j.GetLogger()
		if erro == nil {
			fmt.Fprintf(logger, "can't update status from %s to %s: %s\n", old, j.Status, err)
		}

		return
	}

	return
}

// TODO: undone
func (j *QorJob) GetLogger() (rw io.ReadWriter, err error) {
	if j.log == nil {
		path := fmt.Sprintf("%s/%s-%s-%d.log", WorkerDataPath, j.WorkerName, j.JobName, j.Id)
		j.log, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	rw = j.log

	return
}

func (j *QorJob) GetLog() (l string) {
	log, err := j.GetLogger()
	if err != nil {
		return "failed to retrieve log: " + err.Error()
	}
	lbytes, err := ioutil.ReadAll(log)
	if err != nil {
		return "failed to retrieve log: " + err.Error()
	}
	l = string(lbytes)
	return
}

func (j *QorJob) SavePID() (err error) {
	j.PID = os.Getpid()
	if err = jobDB.Save(j).Error; err != nil {
		logger, erro := j.GetLogger()
		if erro == nil {
			fmt.Fprintf(logger, "can't save pid for job %d\n", j.Id)
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
	return fmt.Sprintf("%s/%s/%d", w.admin.GetRouter().Prefix, utils.ToParamString(w.Name), q.Id)
}

func (j *QorJob) Stop() (err error) { return }

// func (j *QorJob) Kill() (err error)  { return }
func (j *QorJob) Start() (err error) { return }
