package resource

import (
	"github.com/qor/qor"

	"reflect"
)

type Processor struct {
	Result    interface{}
	Resource  *Resource
	Context   *qor.Context
	MetaDatas MetaDatas
}

func (processor *Processor) Validate() (errors []error) {
	resource := processor.Resource
	for _, fc := range resource.validators {
		errors = append(errors, fc(processor.Result, processor.MetaDatas, processor.Context)...)
	}
	return
}

func (processor *Processor) Decode() (errors []error) {
	for _, metaData := range processor.MetaDatas {
		if metaor := metaData.Meta; metaor != nil {
			meta := metaor.GetMeta()
			if len(metaData.MetaDatas) > 0 {
				if resource := meta.GetMeta().Resource; resource != nil {
					field := reflect.Indirect(reflect.ValueOf(processor.Result)).FieldByName(meta.Name)
					if field.Kind() == reflect.Struct {
						association := field.Addr().Interface()
						errors = append(errors, resource.Decode(association, metaData.MetaDatas, processor.Context).Start()...)
					} else if field.Kind() == reflect.Slice {
						value := reflect.New(field.Type().Elem())
						errors = append(errors, resource.Decode(value.Interface(), metaData.MetaDatas, processor.Context).Start()...)
						if !reflect.DeepEqual(reflect.Zero(field.Type().Elem()).Interface(), value.Elem().Interface()) {
							field.Set(reflect.Append(field, value.Elem()))
						}
					}
				}
			} else {
				metaData.Meta.Set(processor.Result, processor.MetaDatas, processor.Context)
			}
		}
	}

	resource := processor.Resource
	for _, fc := range resource.validators {
		errors = append(errors, fc(processor.Result, processor.MetaDatas, processor.Context)...)
	}
	return
}

func (processor *Processor) Commit() (errors []error) {
	errors = processor.Decode()

	resource := processor.Resource
	for _, fc := range resource.processors {
		errors = append(errors, fc(processor.Result, processor.MetaDatas, processor.Context)...)
	}
	return
}

func (processor *Processor) Start() (errors []error) {
	if errors = append(errors, processor.Validate()...); len(errors) == 0 {
		errors = append(errors, processor.Commit()...)
	}
	return
}
