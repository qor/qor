package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

func (res *Resource) Action(action *Action) {
	res.actions[action.Name] = action
}

type Action struct {
	Name   string
	Metas  []string
	Handle func(scope *gorm.DB, context *qor.Context) error
	Inline bool
}
