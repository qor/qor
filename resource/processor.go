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
}

// DecodeToResource decode meta values to resource result
func DecodeToResource(res Resourcer, result interface{}, metaValues *MetaValues, context *qor.Context) *processor {
	return &processor{Resource: res, Result: result, Context: context, MetaValues: metaValues}
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

// Initialize initialize a processor
func (processor *processor) Initialize() error {
	err := processor.Resource.CallFindOne(processor.Result, processor.MetaValues, processor.Context)
	processor.checkSkipLeft(err)
	return err
}

// Validate run validators
func (processor *processor) Validate() error {
	var errs qor.Errors
	if processor.checkSkipLeft() {
		return nil
	}

	for _, v := range processor.Resource.GetResource().Validators {
		if errs.AddError(v.Handler(processor.Result, processor.MetaValues, processor.Context)); !errs.HasError() {
			if processor.checkSkipLeft(errs.GetErrors()...) {
				break
			}
		}
	}
	return errs
}

func (processor *processor) decode() (errs []error) {
	if processor.checkSkipLeft() || processor.MetaValues == nil {
		return
	}

	if destroy := processor.MetaValues.Get("_destroy"); destroy != nil {
		return
	}

	newRecord := true
	scope := &gorm.Scope{Value: processor.Result}
	if primaryField := scope.PrimaryField(); primaryField != nil {
		if !primaryField.IsBlank {
			newRecord = false
		} else {
			for _, metaValue := range processor.MetaValues.Values {
				if metaValue.Meta != nil && metaValue.Meta.GetFieldName() == primaryField.Name {
					if v := utils.ToString(metaValue.Value); v != "" && v != "0" {
						newRecord = false
					}
				}
			}
		}
	}

	for _, metaValue := range processor.MetaValues.Values {
		meta := metaValue.Meta
		if meta == nil {
			continue
		}

		if newRecord && !meta.HasPermission(roles.Create, processor.Context) {
			continue
		} else if !newRecord && !meta.HasPermission(roles.Update, processor.Context) {
			continue
		}

		if setter := meta.GetSetter(); setter != nil {
			setter(processor.Result, metaValue, processor.Context)
		}

		if metaValue.MetaValues != nil && len(metaValue.MetaValues.Values) > 0 {
			if res := metaValue.Meta.GetResource(); res != nil && !reflect.ValueOf(res).IsNil() {
				field := reflect.Indirect(reflect.ValueOf(processor.Result)).FieldByName(meta.GetFieldName())
				// Only decode nested meta value into struct if no Setter defined
				if meta.GetSetter() == nil || reflect.Indirect(field).Type() == utils.ModelType(res.NewStruct()) {
					if _, ok := field.Addr().Interface().(sql.Scanner); !ok {
						decodeMetaValuesToField(res, field, metaValue, processor.Context)
					}
				}
			}
		}
	}

	return
}

// Start start processor
func (processor *processor) Start() error {
	var errs qor.Errors
	processor.Initialize()
	if errs.AddError(processor.Validate()); !errs.HasError() {
		errs.AddError(processor.Commit())
	}
	if errs.HasError() {
		return errs
	}
	return nil
}

// Commit commit data into result
func (processor *processor) Commit() error {
	var errs qor.Errors
	errs.AddError(processor.decode()...)
	if processor.checkSkipLeft(errs.GetErrors()...) {
		return nil
	}

	for _, p := range processor.Resource.GetResource().Processors {
		if err := p.Handler(processor.Result, processor.MetaValues, processor.Context); err != nil {
			if processor.checkSkipLeft(err) {
				break
			}
			errs.AddError(err)
		}
	}
	return errs
}
