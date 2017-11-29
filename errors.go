package qor

import (
	"strings"
)

// Errors is a struct that used to hold errors array
type Errors struct {
	errors []error
}

// Error get formatted error message
func (errs Errors) Error() string {
	var errors []string
	for _, err := range errs.errors {
		errors = append(errors, err.Error())
	}
	return strings.Join(errors, "; ")
}

// AddError add error to Errors struct
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

// HasError return has error or not
func (errs Errors) HasError() bool {
	return len(errs.errors) != 0
}

// GetErrors return error array
func (errs Errors) GetErrors() []error {
	return errs.errors
}

type errorsInterface interface {
	GetErrors() []error
}
