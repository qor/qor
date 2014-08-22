package worker

import "github.com/qor/qor/resource"

type Worker struct {
	*resource.Resource
	Handler   func(job *Job) error
	OnError   func(job *Job)
	OnSuccess func(job *Job)
	OnStart   func(job *Job)
	OnKill    func(job *Job)
	Adapter
}

func New(name string) *Worker {
	return &Worker{}
}

func (worker *Worker) AllowSchedule() {
	worker.RegisterMeta(&resource.Meta{Name: "Schedule", Type: "schedule"})
}

func (worker *Worker) AllMetas() []resource.Meta {
	return []resource.Meta{}
}

func (worker *Worker) RunHandler(job *Job) {
	defer func() {
		if r := recover(); r != nil {
			worker.OnError(job)
		}
	}()

	worker.OnStart(job)
	worker.Handler(job)
	if len(job.Errors) == 0 {
		worker.OnSuccess(job)
	} else {
		worker.OnError(job)
	}
}

func (worker *Worker) UseAdapter(adapter Adapter) {
	worker.Adapter = adapter
}

func (worker *Worker) Listen() {
	worker.Adapter.Listen(worker)
}
