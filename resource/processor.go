package resource

import (
	"errors"

	"github.com/qor/qor"

	"reflect"
)

var (
	ErrProcessorRecordNotFound = errors.New("record not found")
	ErrProcessorSkipLeft       = errors.New("skip left")
)

type Processor struct {
	Result    interface{}
	Resource  *Resource
	Context   *qor.Context
	MetaDatas MetaDatas
	SkipLeft  bool
}

func (processor *Processor) checkSkipLeft(errs ...error) bool {
	if processor.SkipLeft {
		return true
	}

	for _, err := range errs {
		if err == ErrProcessorSkipLeft {
			processor.SkipLeft = true
			break
		}
	}
	return processor.SkipLeft
}

func (processor *Processor) Initialize() error {
	err := ErrProcessorRecordNotFound
	if finder := processor.Resource.Finder; finder != nil {
		err = finder(processor.Result, processor.MetaDatas, processor.Context)
	}
	processor.checkSkipLeft(err)
	return err
}

func (processor *Processor) Validate() (errors []error) {
	if processor.checkSkipLeft() {
		return
	}

	for _, fc := range processor.Resource.validators {
		erres := fc(processor.Result, processor.MetaDatas, processor.Context)
		if processor.checkSkipLeft(erres...) {
			break
		}
		errors = append(errors, erres...)
	}
	return
}

func (processor *Processor) decode() (errors []error) {
	if processor.checkSkipLeft() {
		return
	}

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
	return
}

func (processor *Processor) Commit() (errors []error) {
	errors = processor.decode()
	if processor.checkSkipLeft(errors...) {
		return
	}

	resource := processor.Resource
	for _, fc := range resource.processors {
		erres := fc(processor.Result, processor.MetaDatas, processor.Context)
		if processor.checkSkipLeft(erres...) {
			break
		}
		errors = append(errors, erres...)
	}
	return
}

func (processor *Processor) Start() (errors []error) {
	processor.Initialize()
	if errors = append(errors, processor.Validate()...); len(errors) == 0 {
		errors = append(errors, processor.Commit()...)
	}
	return
}
