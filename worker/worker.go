package worker

import "github.com/qor/qor/resource"

type Worker struct {
	*resource.Resource
	Handler   func(metaValues *resource.MetaValues)
	OnError   func(metaValues *resource.MetaValues)
	OnSuccess func(metaValues *resource.MetaValues)
	OnStart   func(metaValues *resource.MetaValues)
	OnKill    func(metaValues *resource.MetaValues)
	Adapter
}

func New(name string) Worker {
	return Worker{}
}

func (worker Worker) AllowSchedule() {
	worker.RegisterMeta(&resource.Meta{Name: "Schedule", Type: "schedule"})
}

func (worker Worker) AllMetas() []resource.Meta {
	return []resource.Meta{}
}

func (worker Worker) UseAdapter(adapter Adapter) {
	worker.Adapter = adapter
}

func (worker Worker) Listen() {
	worker.Adapter.Listen(worker)
}
