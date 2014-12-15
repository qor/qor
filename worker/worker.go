package worker

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

var (
	jobDB      *gorm.DB
	jobId      uint64
	workerSets []*WorkerSet
	queuers    = map[string]Queuer{}
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
		// w.RunHandler(job)
		for _, ws := range workerSets {
			if ws.Name == job.WokerSetName {
				for _, w := range ws.Workers {
					if w.Name == job.WorkerName {
						w.Run(&job)
					}
				}
			}
		}
		fmt.Fprintf(os.Stderr, "unknown worker(%s:%s) in job(%s)\n", job.WokerSetName, job.WorkerName, job.Id)
		os.Exit(1)
	} else {
		// for _, ws := range workerSets {
		// 	for _, w := range ws.Workers {
		// 		go w.Queuer.Listen(w)
		// 	}
		// }
		for _, queuer := range queuers {
			go func() {
				for {
					job, err := queuer.Dequeue()
					if err != nil {
						// TODO: log
					} else {
						go RunJob(job)
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

	return
}

func (ws *WorkerSet) NewWorker(name string, handle func(job *Job) error, queuer Queuer) (w *Worker) {
	w = &Worker{
		Name:   name,
		Handle: handle,
		Queuer: queuer,
		set:    ws,
	}
	// w.Name = name
	// w.Resource = res
	// w.AllowSchedule()
	// w.Handle = handle
	// w.UseQueuer(queuer)

	ws.Workers = append(ws.Workers, w)

	return
}

type Worker struct {
	// *resource.Resource
	Name   string
	Queuer Queuer

	Config *qor.Config

	set *WorkerSet

	// OnError   func(job *Job, err error)
	Handle    func(job *Job) error
	OnStart   func(job *Job) error
	OnKill    func(job *Job) error
	OnSuccess func(job *Job)
	OnFailed  func(job *Job)
}

// func New(name string, res *resource.Resource) *Worker {
// 	worker := &Worker{}
// 	w.Name = name
// 	// w.Resource = res
// 	w.AllowSchedule()
// 	return worker
// }

// func (w *Worker) AllowSchedule() {
// 	// TODO: how to extend a new meta type
// 	// w.RegisterMeta(&resource.Meta{Name: "Schedule", Type: "schedule"})
// 	// w.RegisterMeta(&resource.Meta{Name: "StartAt", Type: "date"})
// 	// w.RegisterMeta(&resource.Meta{Name: "Interval", Type: "int64"})
// }

// // TODO: to remove? what is this for?
// func (w *Worker) AllMetas() []resource.Meta {
// 	return []resource.Meta{}
// }

func (w *Worker) Run(job *Job) (err error) {
	if w.OnStart != nil {
		if err = w.OnStart(job); err != nil {
			return
		}
	}

	// if err == w.Queuer.Run(job); err != nil {
	// 	if w.OnFailed != nil {
	// 		w.OnFailed(job)
	// 	}
	// 	fmt.Fprintf(os.Stderr, err)
	// 	os.Exit(1)
	// }

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

func (w *Worker) Kill(job *Job) (err error) {
	if w.OnKill != nil {
		if err = w.OnKill(job); err != nil {
			// err = fmt.Fprintf(w.GetLogger(job), "worker.OnKill (%s): %s", w.Name, err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// if err = w.Queuer.Kill(job); err != nil {
	// 	fmt.Fprintf(os.Stderr, err)
	// 	os.Exit(1)
	// }

	return
}

func (w *Worker) NewJob(interval int64, startAt time.Time) (job *Job, err error) {
	job = &Job{
		Interval:     interval,
		StartAt:      startAt,
		WokerSetName: w.set.Name,
		WorkerName:   w.Name,
	}
	if err = jobDB.Save(&job).Error; err != nil {
		return
	}

	w.Queuer.Enqueue(job)

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
