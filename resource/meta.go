package resource

import "github.com/qor/qor/rule"

type Meta struct {
	Name       string
	Type       string
	Label      string
	Value      interface{}
	Collection []Meta
	Resource   interface{}
	Permission rule.Permission
}

type meta struct {
	resource *Resource
	metas    []Meta
}

func (m *meta) Register(meta Meta) {
	m.metas = append(m.metas, meta)
}
