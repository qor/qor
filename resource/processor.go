package resource

import (
	"errors"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

var (
	ErrProcessorRecordNotFound = errors.New("resource: record not found")
	ErrProcessorSkipLeft       = errors.New("resource: skip left")
)

type processor struct {
	Result     interface{}
	Resource   Resourcer
	Context    *qor.Context
	MetaValues *MetaValues
	SkipLeft   bool
	newRecord  bool
}

func DecodeToResource(res Resourcer, result interface{}, metaValues *MetaValues, context *qor.Context) *processor {
	scope := &gorm.Scope{Value: result}
	return &processor{Resource: res, Result: result, Context: context, MetaValues: metaValues, newRecord: scope.PrimaryKeyZero()}
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
	err := processor.Resource.CallFindOne(processor.Result, processor.MetaValues, processor.Context)
	processor.checkSkipLeft(err)
	return err
}

func (processor *processor) Validate() (errors []error) {
	if processor.checkSkipLeft() {
		return
	}

	for _, fc := range processor.Resource.GetResource().validators {
		if err := fc(processor.Result, processor.MetaValues, processor.Context); err != nil {
			if processor.checkSkipLeft(err) {
				break
			}
			errors = append(errors, err)
		}
	}
	return
}

func (processor *processor) decode() (errors []error) {
	if processor.checkSkipLeft() {
		return
	}

	for _, metaValue := range processor.MetaValues.Values {
		meta := metaValue.Meta
		if meta == nil {
			continue
		}

		if processor.newRecord && !meta.HasPermission(roles.Create, processor.Context) {
			continue
		} else if !meta.HasPermission(roles.Update, processor.Context) {
			continue
		}

		if metaValue.MetaValues == nil {
			if setter := meta.GetSetter(); setter != nil {
				setter(processor.Result, metaValue, processor.Context)
			}
			continue
		}

		res := metaValue.Meta.GetResource()
		if res == nil {
			continue
		}

		field := reflect.Indirect(reflect.ValueOf(processor.Result)).FieldByName(meta.GetFieldName())
		if field.Kind() == reflect.Struct {
			association := field.Addr().Interface()
			DecodeToResource(res, association, metaValue.MetaValues, processor.Context).Start()
		} else if field.Kind() == reflect.Slice {
			value := reflect.New(field.Type().Elem())
			associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, processor.Context)
			associationProcessor.Start()
			if !associationProcessor.SkipLeft {
				if !reflect.DeepEqual(reflect.Zero(field.Type().Elem()).Interface(), value.Elem().Interface()) {
					field.Set(reflect.Append(field, value.Elem()))
				}
			}
		}
		processor.MetaValues.Errors = append(processor.MetaValues.Errors, metaValue.MetaValues.Errors...)
	}

	return
}

func (processor *processor) Commit() (errors []error) {
	errors = processor.decode()
	if processor.checkSkipLeft(errors...) {
		return
	}

	for _, fc := range processor.Resource.GetResource().processors {
		if err := fc(processor.Result, processor.MetaValues, processor.Context); err != nil {
			if processor.checkSkipLeft(err) {
				break
			}
			errors = append(errors, err)
		}
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
