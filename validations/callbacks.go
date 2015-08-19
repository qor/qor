package validations

import "github.com/jinzhu/gorm"

var skipValidations = "validations:skip_validations"

func validate(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		if result, ok := scope.DB().Get(skipValidations); !(ok && result.(bool)) {
			scope.CallMethodWithErrorCheck("Validate")
		}
	}
}

func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()
	callback.Create().Before("gorm:before_create").Register("validations:validate", validate)
	callback.Update().Before("gorm:before_update").Register("validations:validate", validate)
}
