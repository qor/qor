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
	jobDB         *gorm.DB
	jobId         uint64
	workerSets    []*WorkerSet
	queuers       = map[string]Queuer{}
	DefaultJobCli = strings.Join(os.Args[0])
)

func init() {
	flag.Uint64Var(&jobId, "job-id", 0, "qor job id")
}

func SetJobDB(db *gorm.DB) {
	jobDB = db
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
		// // w.RunHandler(job)
		// for _, ws := range workerSets {
		// 	if ws.Name == job.WokerSetName {
		// 		for _, w := range ws.Workers {
		// 			if w.Name == job.WorkerName {
		// 				w.Run(&job)
		// 			}
		// 		}
		// 	}
		// }
		// fmt.Fprintf(os.Stderr, "unknown worker(%s:%s) in job(%s)\n", job.WokerSetName, job.WorkerName, job.Id)
		w, err := job.GetWorker()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		w.Run(job)
	} else {
		for _, queuer := range queuers {
			go func() {
				for {
					jobId, err := queuer.Dequeue()
					if err != nil {
						// TODO: log
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

func NewWorkerSet(name string, a *admin.Admin) (ws *WorkerSet) {
	ws = &WorkerSet{Name: name}
	workerSets = append(workerSets, ws)
	a.GetRouter().Get("^/workers$", func(ctx *admin.Context) {})
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

func (w *Worker) Run(job *Job) (err error) {
	if w.OnStart != nil {
		if err = w.OnStart(job); err != nil {
			return
		}
	}

	if err = w.Handle(job); err != nil {
		if w.OnFailed != nil {
			w.OnFailed(job)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if w.OnSuccess != nil {
		w.OnSuccess(job)
	}

	return
}

var ErrJobRun = errors.New("job is already run")

func (w *Worker) Kill(job *Job) (err error) {
	if w.OnKill != nil {
		if err = w.OnKill(job); err != nil {
			// err = fmt.Fprintf(w.GetLogger(job), "worker.OnKill (%s): %s", w.Name, err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	switch job.Status {
	case JobToRun:
		err = w.Queuer.Purge(job)
	case JobRunning:
		// TODO
	case JobRun:
		return ErrJobRun
	}

	return
}

func (w *Worker) NewJob(interval int64, startAt time.Time) (job *Job, err error) {
	// job = &Job{
	// 	Interval:     interval,
	// 	StartAt:      startAt,
	// 	WokerSetName: w.set.Name,
	// 	WorkerName:   w.Name,
	// 	Cli:          DefaultJobCli,
	// }
	// if err = jobDB.Save(&job).Error; err != nil {
	// 	return
	// }

	// err = w.Queuer.Enqueue(job)

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

	return
}

// func (w *Worker) UseQueuer(queuer Queuer) {
// 	w.Queuer = queuer
// }

// func (w *Worker) Listen() {
// 	flag.Parse()
// 	if jobId != "" {
// 		w.RunHandler(job)
// 	} else {
// 		w.Queuer.Listen(worker)
// 	}
// }
