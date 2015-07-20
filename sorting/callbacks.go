package sorting

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
)

func initalizePosition(scope *gorm.Scope) {
	if position, ok := scope.Value.(positionInterface); ok {
		if pos, err := strconv.Atoi(fmt.Sprintf("%v", scope.PrimaryKeyValue())); err == nil {
			if scope.DB().UpdateColumn("position", pos).Error == nil {
				position.SetPosition(pos)
			}
		}
	}
}

func RegisterCallbacks(db *gorm.DB) {
	db.Callback().Create().Before("gorm:commit_or_rollback_transaction").
		Register("sorting:initalize_position", initalizePosition)
}
