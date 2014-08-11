package resource

type MetaValue struct {
	Name       string
	Value      interface{}
	MetaValues MetaValues
	Meta       Metaor
}

type MetaValues []*MetaValue

func (mvs MetaValues) Get(name string) *MetaValue {
	for _, mv := range mvs {
		if mv.Name == name {
			return mv
		}
	}

	return nil
}
