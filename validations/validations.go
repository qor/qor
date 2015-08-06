package validations

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func AddError(db *gorm.DB, resource interface{}, column, err string) {
	scope := db.NewScope(resource)
	var errors = map[string]string{}

	if e, ok := db.Get("validations:errors"); ok {
		errors = e.(map[string]string)
	}

	key := fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), column)
	errors[key] = err

	db.InstantSet("validations:errors", errors)
}

func GetErrors(db *gorm.DB) map[string]string {
	var errors = map[string]string{}
	if e, ok := db.Get("validations:errors"); ok {
		errors = e.(map[string]string)
	}
	return errors
}
