package worker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// type JobStatus string

const (
	// Job statuses
	JobToRun   = "torun"
	JobRunning = "running"
	JobFailed  = "failed"
	JobKilled  = "killed"
	JobRun     = "done"
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

	ExtraInputs string `sql:"type:text;"` // Mysql: 64KB

	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt time.Time

	log *os.File
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
	case JobRunning:
		j.RunCounter++
	case JobFailed:
		j.FailCounter++
	case JobRun:
		j.SuccessCounter++
	case JobKilled:
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

var JobLoggerDir = "/tmp"

// TODO: undone
func (j *QorJob) GetLogger() (rw io.ReadWriter, err error) {
	if j.log == nil {
		path := fmt.Sprintf("%s/%s-%d.log", JobLoggerDir, j.WorkerName, j.JobName, j.Id)
		j.log, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	rw = j.log

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

func (j *QorJob) Stop() (err error) { return }

// func (j *QorJob) Kill() (err error)  { return }
func (j *QorJob) Start() (err error) { return }
