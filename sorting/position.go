package sorting

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/jinzhu/gorm"
)

type positionInterface interface {
	GetPosition() int
	SetPosition(int)
}

type Sorting struct {
	Position int `sql:"DEFAULT:NULL"`
}

func (position Sorting) GetPosition() int {
	return position.Position
}

func (position Sorting) SetPosition(pos int) {
	position.Position = pos
}

func newModel(value interface{}) interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(value)).Type()).Interface()
}

func getSum(a, b interface{}) int {
	ai, _ := strconv.Atoi(fmt.Sprintf("%d", a))
	bi, _ := strconv.Atoi(fmt.Sprintf("%d", b))
	return ai + bi
}

func move(db *gorm.DB, value positionInterface, pos int) error {
	clone := db
	for _, field := range db.NewScope(value).PrimaryFields() {
		if field.DBName != "id" {
			clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), field.Field.Interface())
		}
	}

	if pos > 0 {
		if results := clone.Model(newModel(value)).
			Where("position > ? AND position <= ?", value.GetPosition(), value.GetPosition()+pos).
			UpdateColumn("position", gorm.Expr("position - ?", 1)); results.Error == nil {
			return clone.Model(value).UpdateColumn("position", gorm.Expr("position + ?", results.RowsAffected)).Error
		}
	} else if pos < 0 {
		if results := clone.Model(newModel(value)).
			Where("position < ? AND position >= ?", value.GetPosition(), value.GetPosition()+pos).
			UpdateColumn("position", gorm.Expr("position + ?", 1)); results.Error == nil {
			return clone.Model(value).UpdateColumn("position", gorm.Expr("position - ?", results.RowsAffected)).Error
		}
	}
	return nil
}

func MoveUp(db *gorm.DB, value positionInterface, pos int) error {
	return move(db, value, pos)
}

func MoveDown(db *gorm.DB, value positionInterface, pos int) error {
	return move(db, value, -pos)
}

func MoveTo(db *gorm.DB, value positionInterface, pos int) error {
	return move(db, value, value.GetPosition()-pos)
}
