package audited

import "fmt"

type AuditedModel struct {
	CreatedBy string
	UpdatedBy string
}

func (model *AuditedModel) SetCreatedBy(createdBy interface{}) {
	model.CreatedBy = fmt.Sprintf("%v", createdBy)
}

func (model AuditedModel) GetCreatedBy() string {
	return model.CreatedBy
}

func (model *AuditedModel) SetUpdatedBy(updatedBy interface{}) {
	model.UpdatedBy = fmt.Sprintf("%v", updatedBy)
}

func (model AuditedModel) GetUpdatedBy() string {
	return model.UpdatedBy
}
