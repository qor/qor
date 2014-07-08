package resource

import "github.com/qor/qor/rules"

type Meta struct {
	Name       string
	Type       string
	Label      string
	Value      interface{}
	Collection []Meta
	Resource   interface{}
	Permission rules.Permission
}

type meta struct {
	resource *Resource
	metas    []Meta
}

func (m *meta) Register(meta Meta) {
	m.metas = append(m.metas, meta)
}
