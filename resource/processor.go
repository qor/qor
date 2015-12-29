package resource

import (
	"errors"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/roles"
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

func (processor *processor) Validate() error {
	var errors qor.Errors
	if processor.checkSkipLeft() {
		return nil
	}

	for _, fc := range processor.Resource.GetResource().validators {
		if errors.AddError(fc(processor.Result, processor.MetaValues, processor.Context)); !errors.HasError() {
			if processor.checkSkipLeft(errors.GetErrors()...) {
				break
			}
		}
	}
	return errors
}

func (processor *processor) decode() (errors []error) {
	if processor.checkSkipLeft() || processor.MetaValues == nil {
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

		if setter := meta.GetSetter(); setter != nil {
			setter(processor.Result, metaValue, processor.Context)
			continue
		}

		res := metaValue.Meta.GetResource()
		if res == nil {
			continue
		}

		field := reflect.Indirect(reflect.ValueOf(processor.Result)).FieldByName(meta.GetFieldName())
		if field.Kind() == reflect.Struct {
			value := reflect.New(field.Type())
			associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, processor.Context)
			associationProcessor.Start()
			if !associationProcessor.SkipLeft {
				field.Set(value.Elem())
			}
		} else if field.Kind() == reflect.Slice {
			var fieldType = field.Type().Elem()
			var isPtr bool
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
				isPtr = true
			}

			value := reflect.New(fieldType)
			associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, processor.Context)
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

	return
}

func (processor *processor) Commit() error {
	var errors qor.Errors
	errors.AddError(processor.decode()...)
	if processor.checkSkipLeft(errors.GetErrors()...) {
		return nil
	}

	for _, fc := range processor.Resource.GetResource().processors {
		if err := fc(processor.Result, processor.MetaValues, processor.Context); err != nil {
			if processor.checkSkipLeft(err) {
				break
			}
			errors.AddError(err)
		}
	}
	return errors
}

func (processor *processor) Start() error {
	var errors qor.Errors
	processor.Initialize()
	if errors.AddError(processor.Validate()); !errors.HasError() {
		errors.AddError(processor.Commit())
	}
	if errors.HasError() {
		return errors
	}
	return nil
}
