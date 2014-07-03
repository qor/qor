package resource

type Resource struct {
	attrs *attrs
	meta  *meta
}

func New(value interface{}) *Resource {
	resource := Resource{}
	return &resource
}

func (r *Resource) Attrs() *attrs {
	return r.attrs
}

func (r *Resource) Meta() *meta {
	return r.meta
}
