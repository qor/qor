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

	Cli          string
	WokerSetName string
	WorkerName   string
	Status       string
	PID          int

	log *os.File
}

func (j *Job) GetWorker() (*Worker, error) {
	for _, ws := range workerSets {
		if ws.Name == j.WokerSetName {
			for _, w := range ws.Workers {
				if w.Name == j.WorkerName {
					return w, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("unknown worker(%s:%s) in job(%s)\n", j.WokerSetName, j.WorkerName, j.Id)
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
	j.Status = JobRunning
	if err = jobDB.Model(j).Update("status", j.Status).Error; err != nil {
		logger, erro := j.GetLogger()
		if erro == nil {
			fmt.Fprintf(logger, "can't update status from %s to %s: %s", old, j.Status, err)
		}

		return
	}

	return
}

var JobLoggerDir = "/tmp"

// TODO: undone
func (j *Job) GetLogger() (rw io.ReadWriter, err error) {
	if j.log == nil {
		path := fmt.Sprintf("%s/%s-%s-%d.log", JobLoggerDir, j.WokerSetName, j.WorkerName, j.Id)
		j.log, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	rw = j.log

	return
}

func (j *Job) SavePID() (err error) {
	j.PID = os.Getpid()
	if err = jobDB.Model(j).Update("pid", j.PID).Error; err != nil {
		logger, erro := j.GetLogger()
		if erro == nil {
			fmt.Fprintf(logger, "can't save pid for job %d", j.Id)
		}

		return
	}

	return
}
