package resource

import (
	"database/sql"
	"errors"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

// ErrProcessorSkipLeft skip left processors error, if returned this error in validation, before callbacks, then qor will stop process following processors
var ErrProcessorSkipLeft = errors.New("resource: skip left")

type processor struct {
	Result     interface{}
	Resource   Resourcer
	Context    *qor.Context
	MetaValues *MetaValues
	SkipLeft   bool
	newRecord  bool
}

// DecodeToResource decode meta values to resource result
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

	for _, fc := range processor.Resource.GetResource().Validators {
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

	if destroy := processor.MetaValues.Get("_destroy"); destroy != nil {
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
		}

		if metaValue.MetaValues != nil && len(metaValue.MetaValues.Values) > 0 {
			if res := metaValue.Meta.GetResource(); res != nil && !reflect.ValueOf(res).IsNil() {
				field := reflect.Indirect(reflect.ValueOf(processor.Result)).FieldByName(meta.GetFieldName())
				if utils.ModelType(field.Addr().Interface()) == utils.ModelType(res.NewStruct()) {
					if _, ok := field.Addr().Interface().(sql.Scanner); !ok {
						decodeMetaValuesToField(res, field, metaValue, processor.Context)
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

	for _, fc := range processor.Resource.GetResource().Processors {
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
