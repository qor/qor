package resource

type attrs struct {
	indexAttrs []string
	newAttrs   []string
	editAttrs  []string
	showAttrs  []string
}

func (a *attrs) Index(columns ...string) {
	a.indexAttrs = columns
}

func (a *attrs) New(columns ...string) {
	a.newAttrs = columns
}

func (a *attrs) Edit(columns ...string) {
	a.editAttrs = columns
}

func (a *attrs) Show(columns ...string) {
	a.showAttrs = columns
}

func (resource *Resource) IndexAttrs() (metas []Meta) {
	for _, attr := range resource.attrs.indexAttrs {
		metaDefined := false
		for _, meta := range resource.meta.metas {
			if meta.Name == attr {
				metas = append(metas, meta)
				metaDefined = true
				break
			}
		}
		if !metaDefined {
			metas = append(metas, Meta{Name: attr})
		}
	}
	return
}

func (resource *Resource) NewAttrs() []Meta {
	return resource.IndexAttrs()
}

func (resource *Resource) EditAttrs() []Meta {
	return resource.IndexAttrs()
}

func (resource *Resource) ShowAttrs() []Meta {
	return resource.IndexAttrs()
}
