package validations

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func NewError(resource interface{}, column, err string) error {
	return &Error{Resource: resource, Column: column, Message: err}
}

type Error struct {
	Resource interface{}
	Column   string
	Message  string
}

func (err Error) Label() string {
	scope := gorm.Scope{Value: err.Resource}
	return fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), err.Column)
}

func (err Error) Error() string {
	return fmt.Sprintf("%v", err.Message)
}
