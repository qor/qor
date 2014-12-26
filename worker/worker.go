package worker

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	defaultWorkerSet = &WorkerSet{Name: "qor"}
	workerSets       = []*WorkerSet{defaultWorkerSet}
	queuers          = map[string]Queuer{}
	DefaultJobCli    = strings.Join(os.Args, " ")
)

func init() {
	flag.Uint64Var(&jobId, "job-id", 0, "qor job id")
}

// SetJobDB will run a auto migration for creating table jobs
func SetJobDB(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&Job{}).Error
	if err != nil {
		return
	}

	jobDB = db

	return
}

func NewWorker(queuer Queuer, name string, handle func(job *Job) error) (w *Worker) {
	return defaultWorkerSet.NewWorker(queuer, name, handle)
}

// TODO: UNDONE
func SetAdmin(a *admin.Admin) {
	ws := defaultWorkerSet.Workers
	a.NewResource(&Job{})

	// defaultWorkerSet = NewWorkerSet(defaultWorkerSet.Name, "/workers", "", admin)
	admin.RegisterViewPath(os.Getenv("GOPATH") + "/src/github.com/qor/qor/worker/templates")
	a.GetRouter().Get("/workers", func(c *admin.Context) {
		content := admin.Content{Context: c, Admin: a}
		a.Render("workers", content)
	})
	defaultWorkerSet.Workers = ws
}

// Listen will parse an flag named as "job-id". If the job-id is zero, it
// will run as queue listen server. Otherwise, it will run a specific job
// and terminate the process after the job is run.
//
// It must be executed before http.ListenAndServer
func Listen() {
	flag.Parse()

	if jobId > 0 {
		var job Job
		if err := jobDB.Where("id = ?", jobId).Find(&job).Error; err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var logger io.Writer
		var err error
		if logger, err = job.GetLogger(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		var w *Worker
		if w, err = job.GetWorker(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		if err = w.Run(&job); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		os.Exit(0)
	} else {
		for _, queuer := range queuers {
			go func() {
				for {
					jobId, err := queuer.Dequeue()
					fmt.Println("dequeue job", jobId)
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

// TODO: UNDONE
func NewWorkerSet(name, router, tmplDir string, a *admin.Admin, handle func(ctx *admin.Context)) (ws *WorkerSet) {
	ws = &WorkerSet{Name: name}
	workerSets = append(workerSets, ws)
	a.GetRouter().Get(router, handle)
	// template register
	// menu register
	// Job resource register

	return
}

func (ws *WorkerSet) NewWorker(queuer Queuer, name string, handle func(job *Job) error) (w *Worker) {
	w = &Worker{
		Name:   name,
		Handle: handle,
		Queuer: queuer,
		set:    ws,
	}

	ws.Workers = append(ws.Workers, w)

	queuers[queuer.Name()] = queuer

	return
}

type Worker struct {
	Name   string
	Queuer Queuer
	Config *qor.Config
	set    *WorkerSet

	Handle func(job *Job) error
	OnKill func(job *Job) error
	// OnStart   func(job *Job) error
	// OnSuccess func(job *Job)
	// OnFailed  func(job *Job)
}

func (w *Worker) Run(job *Job) (err error) {
	if err = job.SavePID(); err != nil {
		return
	}
	logger, err := job.GetLogger()
	if err != nil {
		return
	}

	fmt.Fprintf(logger, "run job (%d) with pid (%d)\n", job.Id, job.PID)

	if err = job.UpdateStatus(JobRunning); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		return
	}

	// if w.OnStart != nil {
	// 	if err = w.OnStart(job); err != nil {
	// 		logger.Write([]byte("worker.onstart: " + err.Error() + "\n"))

	// 		if err = job.UpdateStatus(JobFailed); err != nil {
	// 			fmt.Fprintf(logger, "error: %s\n", err)
	// 			os.Exit(1)
	// 		}

	// 		fmt.Fprintf(logger, "error: %s\n", err)
	// 		os.Exit(1)
	// 	}
	// }

	if err = w.Handle(job); err != nil {
		logger.Write([]byte("worker.hanlde: " + err.Error() + "\n"))

		if err = job.UpdateStatus(JobFailed); err != nil {
			fmt.Fprintf(logger, "error: %s\n", err)
			os.Exit(1)
		}

		// if w.OnFailed != nil {
		// 	w.OnFailed(job)
		// }

		// } else if w.OnSuccess != nil {
		// 	if err = job.UpdateStatus(JobRun); err != nil {
		// 		fmt.Fprintf(logger, "error: %s\n", err)
		// 	}

		// 	w.OnSuccess(job)
	}

	if err = job.UpdateStatus(JobRun); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(logger, "finish job (%d) with pid (%d)\n", job.Id, job.PID)

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

		err = process.Kill()
	case JobRun:
		return ErrJobRun
	}

	if err == nil {
		err = job.UpdateStatus(JobKilled)
	}

	return
}

func (w *Worker) NewJob(interval uint64, startAt time.Time) (job *Job, err error) {
	return w.NewJobWithCli(interval, startAt, DefaultJobCli)
}

func (w *Worker) NewJobWithCli(interval uint64, startAt time.Time, cli string) (job *Job, err error) {
	job = &Job{
		Interval:     interval,
		StartAt:      startAt,
		WokerSetName: w.set.Name,
		WorkerName:   w.Name,
		Cli:          cli,
	}
	if err = jobDB.Save(job).Error; err != nil {
		return
	}

	err = w.Queuer.Enqueue(job)

	if job.QueueJobId != "" {
		if err = jobDB.Save(job).Error; err != nil {
			return
		}
	}

	return
}
