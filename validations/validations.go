package validations

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

var settingKey = "validations:errors"

func AddError(db *gorm.DB, resource interface{}, err string) {
	var validationErrors = GetErrors(db)
	var scope = db.NewScope(resource)

	key := fmt.Sprintf("%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue())
	validationErrors[key] = append(validationErrors[key], err)

	db.InstantSet(settingKey, validationErrors).Error = fmt.Errorf("RecordInvalid: %v", err)
}

func AddErrorForColumn(db *gorm.DB, resource interface{}, column, err string) {
	var validationErrors = GetErrors(db)
	var scope = db.NewScope(resource)

	key := fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), column)
	validationErrors[key] = append(validationErrors[key], err)

	db.InstantSet(settingKey, validationErrors).Error = fmt.Errorf("RecordInvalid: %v", err)
}

func GetErrors(db *gorm.DB) map[string][]string {
	var validationErrors = map[string][]string{}
	if errors, ok := db.Get(settingKey); ok {
		validationErrors = errors.(map[string][]string)
	}
	return validationErrors
}
