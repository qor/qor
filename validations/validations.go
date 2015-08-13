package validations

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type Error struct {
	Column  string
	Message string
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Column, err.Message)
}

func NewError(db *gorm.DB, resource interface{}, column, err string) Error {
	scope := db.NewScope(resource)
	key := fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), column)
	return Error{Column: key, Message: err}
}
