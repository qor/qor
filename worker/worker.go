package worker

import (
	"github.com/qor/qor/admin"
	"flag"
	"time"

	"github.com/qor/qor/resource"
)

var jobId uint64

func init() {
	flag.IntVar(&jobId, "job-id", 0, "qor job id")
}

type Worker struct {
	// *resource.Resource
	StartAt   time.Time
	Name      string
	Interval  int
	Handler   func(job *Job) error
	OnError   func(job *Job)
	OnSuccess func(job *Job)
	OnStart   func(job *Job)
	OnKill    func(job *Job)
	Adapter
}

type WorkerSet []*Worker

func NewWorkerSet(router *admin.Router) (ws WorkerSet) {
	router.Get("^/workers$", func(ctx *admin.Context) {

	})
	// router.Get("^/workers/.*$", admin.WorkersAction)
}

func (ws *WorkerSet) Listen() {

}

func (ws *WorkerSet) NewWorker(name string) *Worker {
	ws = append(ws, New(name))
}

func New(name string, res *resource.Resource) *Worker {
	worker := &Worker{}
	worker.Name = name
	// worker.Resource = res
	worker.AllowSchedule()
	return worker
}

func (worker *Worker) AllowSchedule() {
	// TODO: how to extend a new meta type
	// worker.RegisterMeta(&resource.Meta{Name: "Schedule", Type: "schedule"})
	// worker.RegisterMeta(&resource.Meta{Name: "StartAt", Type: "date"})
	// worker.RegisterMeta(&resource.Meta{Name: "Interval", Type: "int64"})
}

// TODO: to remove? what is this for?
func (worker *Worker) AllMetas() []resource.Meta {
	return []resource.Meta{}
}

func (worker *Worker) RunHandler(job *Job) {
	defer func() {
		if r := recover(); r != nil {
			if worker.OnError != nil {
				worker.OnError(job)
			}
		}
	}()

	if worker.OnStart != nil {
		worker.OnStart(job)
	}

	worker.Handler(job)

	if len(job.Errors) == 0 {
		if worker.OnSuccess != nil {
			worker.OnSuccess(job)
		}
	} else {
		if worker.OnError != nil {
			worker.OnError(job)
		}
	}
}

func (worker *Worker) UseAdapter(adapter Adapter) {
	worker.Adapter = adapter
}

func (worker *Worker) Listen() {
	flag.Parse()
	if jobId != "" {
		worker.RunHandler(job)
	} else {
		worker.Adapter.Listen(worker)
	}
}

func Listen() {
	flag.Parse()
	if jobId > 0 {
		worker.RunHandler(job)
	} else {
	}
}
