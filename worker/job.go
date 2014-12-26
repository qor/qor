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

type Job struct {
	Id         uint64
	QueueJobId string

	// unit: minute
	// 0 to run job only once
	Interval uint64

	// zero time value to execute job immediately
	StartAt time.Time

	Cli        string
	WorkerName string
	Status     string
	PID        int

	// RunCounter uint64

	log *os.File
}

func (j *Job) GetWorker() (w *Worker, err error) {
	var ok bool
	if w, ok = workers[j.WorkerName]; !ok {
		err = fmt.Errorf("unknown worker: %s\n", j.WorkerName)
	}

	return
}

func RunJob(jobId uint64) {
	job := &Job{}
	if err := jobDB.Find(job, jobId).Error; err != nil {
		fmt.Printf("job (%d) do not existed\n", jobId)
	} else {
		job.Run()
	}
}

func (j *Job) Run() (err error) {
	parts := strings.Split(j.Cli, " ")
	name := parts[0]
	args := []string{"-job-id", strconv.FormatUint(j.Id, 10)}
	if len(parts) > 1 {
		args = append(args, parts[1:]...)
	}

	err = exec.Command(name, args...).Start()
	return
}

func (j *Job) UpdateStatus(status string) (err error) {
	old := j.Status
	j.Status = status
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
func (j *Job) GetLogger() (rw io.ReadWriter, err error) {
	if j.log == nil {
		path := fmt.Sprintf("%s/%s-%d.log", JobLoggerDir, j.WorkerName, j.Id)
		j.log, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	rw = j.log

	return
}

func (j *Job) SavePID() (err error) {
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

func (j *Job) Stop() (err error) { return }

// func (j *Job) Kill() (err error)  { return }
func (j *Job) Start() (err error) { return }
