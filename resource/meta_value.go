package resource

import (
	"reflect"

	"github.com/qor/qor"
)

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
	Index      int
	Meta       Metaor
	MetaValues *MetaValues
}

func decodeMetaValuesToField(res Resourcer, field reflect.Value, metaValue *MetaValue, context *qor.Context) {
	if field.Kind() == reflect.Struct {
		value := reflect.New(field.Type())
		associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, context)
		associationProcessor.Start()
		if !associationProcessor.SkipLeft {
			field.Set(value.Elem())
		}
	} else if field.Kind() == reflect.Slice {
		if metaValue.Index == 0 {
			field.Set(reflect.Zero(field.Type()))
		}

		var fieldType = field.Type().Elem()
		var isPtr bool
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtr = true
		}

		value := reflect.New(fieldType)
		associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, context)
		associationProcessor.Start()
		if !associationProcessor.SkipLeft {
			if !reflect.DeepEqual(reflect.Zero(fieldType).Interface(), value.Elem().Interface()) {
				if isPtr {
					field.Set(reflect.Append(field, value))
				} else {
					field.Set(reflect.Append(field, value.Elem()))
				}
			}
		}
	}
}
