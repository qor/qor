package worker

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

var (
	jobDB            *gorm.DB
	jobId            uint64
	defaultWorkerSet = &WorkerSet{Name: "default"}
	workerSets       = []*WorkerSet{defaultWorkerSet}
	queuers          = map[string]Queuer{}
	DefaultJobCli    = strings.Join(os.Args, " ")
)

func init() {
	flag.Uint64Var(&jobId, "job-id", 0, "qor job id")
}

func SetJobDB(db *gorm.DB) {
	jobDB = db
}

func NewWorker(name string, handle func(job *Job) error, queuer Queuer) (w *Worker) {
	return defaultWorkerSet.NewWorker(name, handle, queuer)
}

func SetAdmin(admin *admin.Admin) {
	ws := defaultWorkerSet.Workers
	defaultWorkerSet = NewWorkerSet(defaultWorkerSet.Name, "/workers", "", admin)
	defaultWorkerSet.Workers = ws
}

func RegisterQueuer(name string, queuer Queuer) {
	queuers[name] = queuer
}

func Listen() {
	flag.Parse()

	if jobId > 0 {
		var job Job
		if err := jobDB.Where("_id = ?", jobId).Find(&job).Error; err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		w, err := job.GetWorker()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		w.Run(&job)
	} else {
		for _, queuer := range queuers {
			go func() {
				for {
					jobId, err := queuer.Dequeue()
					if err != nil {
						fmt.Println("qor.worker.dequeue.error:", err)
					} else {
						go RunJob(jobId)
					}
				}
			}()
		}
	}
}

type WorkerSet struct {
	Name    string
	Workers []*Worker
}

func NewWorkerSet(name, router, tmplDir string, a *admin.Admin) (ws *WorkerSet) {
	ws = &WorkerSet{Name: name}
	workerSets = append(workerSets, ws)
	a.GetRouter().Get(router, func(ctx *admin.Context) {})
	// template register
	// menu register
	// Job resource register

	return
}

func (ws *WorkerSet) NewWorker(name string, handle func(job *Job) error, queuer Queuer) (w *Worker) {
	w = &Worker{
		Name:   name,
		Handle: handle,
		Queuer: queuer,
		set:    ws,
	}

	ws.Workers = append(ws.Workers, w)

	return
}

type Worker struct {
	Name   string
	Queuer Queuer
	Config *qor.Config
	set    *WorkerSet

	Handle    func(job *Job) error
	OnStart   func(job *Job) error
	OnKill    func(job *Job) error
	OnSuccess func(job *Job)
	OnFailed  func(job *Job)
}

// TODO: use docker
func (w *Worker) Run(job *Job) (err error) {
	if err = job.SavePID(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logger, err := job.GetLogger()
	if err != nil {
		fmt.Println("can't get job logger")
		return
	}

	fmt.Fprintf(logger, "to run job (%d) with pid (%d)", job.Id, job.PID)

	if err = job.UpdateStatus(JobRunning); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if w.OnStart != nil {
		if err = w.OnStart(job); err != nil {
			logger.Write([]byte("worker.onstart: " + err.Error() + "\n"))

			if err = job.UpdateStatus(JobFailed); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err = w.Handle(job); err != nil {
		if err = job.UpdateStatus(JobFailed); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if w.OnFailed != nil {
			w.OnFailed(job)
		}

		logger.Write([]byte("worker.hanlde: " + err.Error() + "\n"))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if w.OnSuccess != nil {
		if err = job.UpdateStatus(JobRun); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		w.OnSuccess(job)
	}

	fmt.Fprintf(logger, "finish job (%d) with pid (%d)", job.Id, job.PID)

	return
}

var ErrJobRun = errors.New("job is already run")

func (w *Worker) Kill(job *Job) (err error) {
	if w.OnKill != nil {
		if err = w.OnKill(job); err != nil {
			return
		}
	}

	switch job.Status {
	case JobToRun:
		err = w.Queuer.Purge(job)
	case JobRunning:
		if job.PID == 0 {
			return errors.New("pid is zero")
		}

		var process *os.Process
		process, err = os.FindProcess(job.PID)
		if err != nil {
			return
		}

		// err = process.Signal(syscall.SIGUSR1)
		err = process.Kill()
	case JobRun:
		return ErrJobRun
	}

	if err == nil {
		err = job.UpdateStatus(JobKilled)
	}

	return
}

func (w *Worker) NewJob(interval int64, startAt time.Time) (job *Job, err error) {
	return w.NewJobWithCli(interval, startAt, DefaultJobCli)
}

func (w *Worker) NewJobWithCli(interval int64, startAt time.Time, cli string) (job *Job, err error) {
	job = &Job{
		Interval:     interval,
		StartAt:      startAt,
		WokerSetName: w.set.Name,
		WorkerName:   w.Name,
		Cli:          DefaultJobCli,
	}
	if err = jobDB.Save(&job).Error; err != nil {
		return
	}

	err = w.Queuer.Enqueue(job)

	if job.QueueJobId != "" {
		if err = jobDB.Save(&job).Error; err != nil {
			return
		}
	}

	return
}
