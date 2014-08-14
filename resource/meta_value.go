package resource

type MetaValue struct {
	Name       string
	Value      interface{}
	MetaValues MetaValues
	Meta       Metaor

	error error
}

type MetaValues struct {
	Values []*MetaValue
	Errors []error
}

func (mvs *MetaValues) AddError(mv MetaValue, err error) {
	mv.error = err
	mvs.Errors = append(mvs.Errors, err)
}

func (mvs MetaValues) Get(name string) *MetaValue {
	for _, mv := range mvs.Values {
		if mv.Name == name {
			return mv
		}
	}

	return nil
}
