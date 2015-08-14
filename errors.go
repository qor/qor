package qor

import "strings"

type Errors struct {
	errors []error
}

func (errs Errors) Error() string {
	var errors []string
	for _, err := range errs.errors {
		errors = append(errors, err.Error())
	}
	return strings.Join(errors, "; ")
}

type errorsInterface interface {
	GetErrors() []error
}

func (errs *Errors) AddError(errors ...error) {
	for _, err := range errors {
		if err != nil {
			if e, ok := err.(errorsInterface); ok {
				errs.errors = append(errs.errors, e.GetErrors()...)
			} else {
				errs.errors = append(errs.errors, err)
			}
		}
	}
}

func (errs Errors) HasError() bool {
	return len(errs.errors) != 0
}

func (errs Errors) GetErrors() []error {
	return errs.errors
}
