package resource

type MetaValue struct {
	Name       string
	Value      interface{}
	MetaValues *MetaValues
	Meta       Metaor
	error      error
}

type MetaValues struct {
	Values []*MetaValue
}

func (mvs MetaValues) Get(name string) *MetaValue {
	for _, mv := range mvs.Values {
		if mv.Name == name {
			return mv
		}
	}

	return nil
}
