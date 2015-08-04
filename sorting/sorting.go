package sorting

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/jinzhu/gorm"
)

type positionInterface interface {
	GetPosition() *int
	SetPosition(int)
}

type Sorting struct {
	Position *int `sql:"DEFAULT:NULL"`
}

func (position Sorting) GetPosition() *int {
	return position.Position
}

func (position *Sorting) SetPosition(pos int) {
	position.Position = &pos
}

func newModel(value interface{}) interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(value)).Type()).Interface()
}

func getRealPosition(db *gorm.DB, value positionInterface) int {
	var currentPos int
	if pos := value.GetPosition(); pos == nil {
		currentPos, _ = strconv.Atoi(fmt.Sprintf("%v", db.NewScope(value).PrimaryKeyValue()))
	} else {
		currentPos = *pos
	}
	return currentPos
}

func move(db *gorm.DB, value positionInterface, pos int) error {
	clone := db
	for _, field := range db.NewScope(value).PrimaryFields() {
		if field.DBName != "id" {
			clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), field.Field.Interface())
		}
	}

	currentPos := getRealPosition(db, value)
	value.SetPosition(currentPos + pos)

	if pos > 0 {
		if results := clone.Model(newModel(value)).
			Where("position > ? AND position <= ?", currentPos, currentPos+pos).
			UpdateColumn("position", gorm.Expr("position - ?", 1)); results.Error == nil {
			return clone.Model(value).UpdateColumn("position", gorm.Expr("position + ?", results.RowsAffected)).Error
		}
	} else if pos < 0 {
		if results := clone.Model(newModel(value)).
			Where("position < ? AND position >= ?", currentPos, currentPos+pos).
			UpdateColumn("position", gorm.Expr("position + ?", 1)); results.Error == nil {
			return clone.Model(value).UpdateColumn("position", gorm.Expr("position - ?", results.RowsAffected)).Error
		}
	}
	return nil
}

func MoveUp(db *gorm.DB, value positionInterface, pos int) error {
	return move(db, value, -pos)
}

func MoveDown(db *gorm.DB, value positionInterface, pos int) error {
	return move(db, value, pos)
}

func MoveTo(db *gorm.DB, value positionInterface, pos int) error {
	return move(db, value, pos-getRealPosition(db, value))
}
