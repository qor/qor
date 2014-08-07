package resource

import "github.com/qor/qor"

type Processor struct {
	Result    interface{}
	Resource  *Resource
	Context   qor.Context
	MetaDatas MetaDatas
}

func (processor *Processor) Validate() (errors []error) {
	resource := processor.Resource
	for _, fc := range resource.validators {
		errors = append(errors, fc(processor.Result, processor.MetaDatas, processor.Context)...)
	}
	return
}

func (processor *Processor) Commit() (errors []error) {
	resource := processor.Resource
	for _, fc := range resource.processors {
		errors = append(errors, fc(processor.Result, processor.MetaDatas, processor.Context)...)
	}
	return
}
