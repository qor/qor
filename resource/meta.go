package resource

type Meta struct {
	Name       string
	Type       string
	Label      string
	Value      interface{}
	Collection interface{}
	Resource   interface{}
}

type meta struct {
	resource *Resource
	metas    []Meta
}

func (m *meta) Register(meta Meta) {
	m.metas = append(m.metas, meta)
}
