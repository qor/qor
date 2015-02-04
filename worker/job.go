package worker

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/qor/qor/admin"
	"github.com/qor/qor/utils"
)

type Job struct {
	Name   string
	Queuer Queuer
	// Config *qor.Config

	Description string

	worker   *Worker
	Resource *admin.Resource
	metas    []*admin.Meta

	Handle    func(job *QorJob) error
	OnKill    func(job *QorJob) error
	OnStart   func(job *QorJob) error
	OnSuccess func(job *QorJob)
	OnFailed  func(job *QorJob)
}

// func (j *Job) ExtraInput(res *admin.Resource) {
// 	j.Resource = res
// }

func (j *Job) Run(job *QorJob) (err error) {
	if err = job.SavePID(); err != nil {
		return
	}
	logger, err := job.GetLogger()
	if err != nil {
		return
	}

	fmt.Fprintf(logger, "run job (%d) with pid (%d)\n", job.Id, job.PID)

	if j.OnStart != nil {
		if err = j.OnStart(job); err != nil {
			logger.Write([]byte("worker.onstart: " + err.Error() + "\n"))

			if err = job.UpdateStatus(JobFailed); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err = job.UpdateStatus(JobRunning); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		return
	}

	if err = j.Handle(job); err != nil {
		if err = job.UpdateStatus(JobFailed); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if j.OnFailed != nil {
			j.OnFailed(job)
		}

		logger.Write([]byte("worker.hanlde: " + err.Error() + "\n"))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if j.OnSuccess != nil {
		if err = job.UpdateStatus(JobRun); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		j.OnSuccess(job)
	}

	if err = job.UpdateStatus(JobRun); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(logger, "finish job (%d) with pid (%d)\n", job.Id, job.PID)

	return
}

var ErrJobRun = errors.New("job is already run")

func (j *Job) Kill(job *QorJob) (err error) {
	if j.OnKill != nil {
		if err = j.OnKill(job); err != nil {
			return
		}
	}

	switch job.Status {
	case JobToRun:
		err = j.Queuer.Purge(job)
	case JobRunning:
		if job.PID == 0 {
			return errors.New("pid is zero")
		}

		var process *os.Process
		process, err = os.FindProcess(job.PID)
		if err != nil {
			return
		}

		err = process.Kill()
	case JobRun:
		return ErrJobRun
	}

	if err == nil {
		err = job.UpdateStatus(JobKilled)
	}

	return
}

func (j *Job) NewQorJob(interval uint64, startAt time.Time) (job *QorJob, err error) {
	return j.NewQorJobWithCli(interval, startAt, DefaultJobCli)
}

func (j *Job) NewQorJobWithCli(interval uint64, startAt time.Time, cli string) (job *QorJob, err error) {
	job = &QorJob{
		Interval:   interval,
		StartAt:    startAt,
		JobName:    j.Name,
		WorkerName: j.worker.Name,
		Cli:        cli,
	}

	if err = jobDB.Save(job).Error; err != nil {
		return
	}

	err = j.Queuer.Enqueue(job)

	if job.QueueJobId != "" {
		if err = jobDB.Save(job).Error; err != nil {
			return
		}
	}

	return
}

func (j *Job) initResource() {
	qorjob := j.worker.admin.AddResource(&QorJob{}, &admin.Config{Name: j.Name + "-QorJob", Invisible: true})
	// qorjob.IndexAttrs("Id", "QueueJobId", "Interval", "StartAt", "Cli", "WorkerName", "Status", "PID", "RunCounter", "FailCounter", "SuccessCounter", "KillCounter")
	qorjob.NewAttrs("Interval", "StartAt")

	// qorjob.Meta(&admin.Meta{Name: "WorkerName", Type: "select_one", Collection: func(interface{}, *qor.Context) [][]string {
	// 	var keys [][]string
	// 	for k, _ := range w.jobs {
	// 		keys = append(keys, []string{k, k})
	// 	}
	// 	return keys
	// }})

	j.Resource = qorjob
}

func (j *Job) Meta(meta *admin.Meta) {
	j.Resource.Meta(meta)
	j.metas = append(j.metas, meta)
	attrs := []string{"Interval", "StartAt"}
	for _, meta := range j.metas {
		attrs = append(attrs, meta.GetName())
	}
	j.Resource.NewAttrs(attrs...)
}

func (j *Job) URL() string {
	return fmt.Sprintf("%s/%s/new?job=%s", j.worker.admin.GetRouter().Prefix, utils.ToParamString(j.worker.Name), j.Name)
}
