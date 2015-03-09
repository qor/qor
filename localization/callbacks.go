package localization

import (
	"reflect"

	"github.com/jinzhu/gorm"
)

func BeforeQuery(scope *gorm.Scope) {
	if _, isLocalization := reflect.New(scope.GetModelStruct().ModelType).Interface().(Interface); isLocalization {
		if str, ok := scope.DB().Get("localization:locale"); ok {
			if locale, ok := str.(string); ok {
				switch mode, _ := scope.DB().Get("localization"); mode {
				case "locale":
					scope.Search.Where("language_code = ?", locale)
				case "global":
					scope.Search.Where("language_code IS NULL")
				default:
					scope.Search.Where("language_code = ? OR language_code IS NULL", locale)
				}
			}
		}
	}
}
