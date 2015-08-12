package audited

import (
	"fmt"

	"github.com/qor/qor/admin"
)

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

// type Audited struct {
// 	gorm.Model
// 	ReferTable    string
// 	ReferId       string
// 	Action        string
// 	ChangeDetails string `sql:"size:65532"`
// 	Comment       string `sql:"size:1024"`
// }

func (model *AuditedModel) InjectQorAdmin(res *admin.Resource) {
	// Middleware
	res.GetAdmin().GetRouter().Use(func(context *admin.Context, middleware *admin.Middleware) {
		context.SetDB(context.GetDB().Set("audited:current_user", context.CurrentUser))
		middleware.Next(context)
	})
}
