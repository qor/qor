package resource

import (
	"errors"
	"fmt"

	"github.com/qor/qor"

	"reflect"
)

var (
	ErrProcessorRecordNotFound = errors.New("record not found")
	ErrProcessorSkipLeft       = errors.New("skip left")
)

type processor struct {
	Result    interface{}
	Resource  Resourcer
	Context   *qor.Context
	MetaDatas MetaDatas
	SkipLeft  bool
}

func DecodeToResource(res Resourcer, result interface{}, metaDatas MetaDatas, context *qor.Context) *processor {
	return &processor{Resource: res, Result: result, Context: context, MetaDatas: metaDatas}
}

func (processor *processor) checkSkipLeft(errs ...error) bool {
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

func (processor *processor) Initialize() error {
	err := ErrProcessorRecordNotFound
	if finder := processor.Resource.GetFinder(); finder != nil {
		fmt.Println(finder)
		err = finder(processor.Result, processor.MetaDatas, processor.Context)
	}
	processor.checkSkipLeft(err)
	return err
}

func (processor *processor) Validate() (errors []error) {
	if processor.checkSkipLeft() {
		return
	}

	for _, fc := range processor.Resource.GetResource().validators {
		erres := fc(processor.Result, processor.MetaDatas, processor.Context)
		if processor.checkSkipLeft(erres...) {
			break
		}
		errors = append(errors, erres...)
	}
	return
}

func (processor *processor) decode() (errors []error) {
	if processor.checkSkipLeft() {
		return
	}

	for _, metaData := range processor.MetaDatas {
		if metaor := metaData.Meta; metaor != nil {
			meta := metaor.GetMeta()
			if len(metaData.MetaDatas) > 0 {
				if res := meta.GetMeta().Resource; res != nil {
					field := reflect.Indirect(reflect.ValueOf(processor.Result)).FieldByName(meta.Name)
					if field.Kind() == reflect.Struct {
						association := field.Addr().Interface()
						errors = append(errors, DecodeToResource(res, association, metaData.MetaDatas, processor.Context).Start()...)
					} else if field.Kind() == reflect.Slice {
						value := reflect.New(field.Type().Elem())
						errors = append(errors, DecodeToResource(res, value.Interface(), metaData.MetaDatas, processor.Context).Start()...)
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

func (processor *processor) Commit() (errors []error) {
	errors = processor.decode()
	if processor.checkSkipLeft(errors...) {
		return
	}

	for _, fc := range processor.Resource.GetResource().processors {
		erres := fc(processor.Result, processor.MetaDatas, processor.Context)
		if processor.checkSkipLeft(erres...) {
			break
		}
		errors = append(errors, erres...)
	}
	return
}

func (processor *processor) Start() (errors []error) {
	processor.Initialize()
	if errors = append(errors, processor.Validate()...); len(errors) == 0 {
		errors = append(errors, processor.Commit()...)
	}
	return
}
