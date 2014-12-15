package worker

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// type JobStatus string

const (
	// Job statuses
	JobToRun   = "hold"
	JobRunning = "running"
	JobRun     = "done"
)

type Job struct {
	Id uint64

	// unit: minute
	// 0 to run job only once
	Interval int64

	// zero time value to execute job immediately
	StartAt time.Time

	Cli string

	WokerSetName string
	WorkerName   string

	Status string
}

func (j *Job) GetWorker() (*Worker, error) {
	for _, ws := range workerSets {
		if ws.Name == job.WokerSetName {
			for _, w := range ws.Workers {
				if w.Name == job.WorkerName {
					// w.Run(&job)
					return w, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("unknown worker(%s:%s) in job(%s)\n", job.WokerSetName, job.WorkerName, job.Id)
}

func RunJob(jobId uint64) {
	job := &Job{}
	if err := jobDB.Find(job, jobId).Error; err != nil {
		// TODO
	} else {
		job.Run()
	}

}

func (j *Job) Run() (err error) {
	parts := strings.Split(j.Cli, " ")
	name := parts[0]
	args := []string{"-job-id", strconv.Itoa(j.Id)}
	if len(parts) > 1 {
		args = append(args, parts[1:]...)
	}

	err = exec.Command(name, args...).Start()
	return
}

func (j *Job) Stop() (err error) {
	return
}

func (j *Job) GetLogger() io.ReadWriter {

}

// func (job *Job) AddErr(err error) {
// 	job.Errors = append(job.Errors, err)
// }

// func (job *Job) GetProcessLog() io.Reader {
// 	return job.Worker.GetProcessLog(job)
// }

// func (job *Job) LogWriter() io.Writer {
// 	return job.Worker.LogWriter(job)
// }

// func (job *Job) Kill() {
// 	if job.Worker.Kill(job) {
// 		job.Worker.OnKill(job)
// 	}
// }
