package resource

import "github.com/qor/qor"

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
		if metaData.Meta != nil {
			metaData.Meta.Set(processor.Result, processor.MetaDatas, processor.Context)
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
