package resource

type Resource struct {
	attrs      *attrs
	meta       *meta
	IndexAttrs []string
	ShowAttrs  []string
	NewAttrs   []string
	EditAttrs  []string
}

func New() *Resource {
	resource := Resource{}
	return &resource
}

func (s *Resource) Attrs() *attrs {
	return s.attrs
}

func (s *Resource) Meta() *meta {
	return s.meta
}

func (s *Resource) Search() {
}

func (s *Resource) Filter() {
}

func (s *Resource) Action() {
}

func (s *Resource) Download() {
}
