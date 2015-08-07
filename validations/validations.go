package validations

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

var settingKey = "validations:errors"

func AddError(db *gorm.DB, resource interface{}, err string) {
	scope := db.NewScope(resource)
	var errors = map[string][]string{}

	if e, ok := db.Get(settingKey); ok {
		errors = e.(map[string][]string)
	}

	key := fmt.Sprintf("%v::%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue())
	errors[key] = append(errors[key], err)

	db.InstantSet(settingKey, errors)
}

func AddErrorForColumn(db *gorm.DB, resource interface{}, column, err string) {
}

func GetErrors(db *gorm.DB) map[string][]string {
	var errors = map[string][]string{}
	if e, ok := db.Get(settingKey); ok {
		errors = e.(map[string][]string)
	}
	return errors
}
