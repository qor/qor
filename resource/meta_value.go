package resource

// MetaValues is slice of MetaValue
type MetaValues struct {
	Values []*MetaValue
}

// Get get meta value from MetaValues with name
func (mvs MetaValues) Get(name string) *MetaValue {
	for _, mv := range mvs.Values {
		if mv.Name == name {
			return mv
		}
	}

	return nil
}

// MetaValue a struct used to hold information when convert inputs from HTTP form, JSON, CSV files and so on to meta values
// It will includes field name, field value and its configured Meta, if it is a nested resource, will includes nested metas in its MetaValues
type MetaValue struct {
	Name       string
	Value      interface{}
	MetaValues *MetaValues
	Meta       Metaor
	error      error
}
