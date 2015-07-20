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

func MoveUp(db *gorm.DB, value positionInterface, pos int) error {
	if pos == 0 {
		return nil
	}

	clone := db
	for _, field := range db.NewScope(value).PrimaryFields() {
		if field.DBName != "id" {
			clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), field.Field.Interface())
		}
	}

	if err := clone.Model(newModel(value)).
		Where("position > ? AND position <= ?", value.GetPosition(), gorm.Expr("? + ?", value.GetPosition(), pos)).
		UpdateColumn("position", gorm.Expr("position - ?", 1)).Error; err == nil {
		return clone.Model(value).UpdateColumn("position", gorm.Expr("position + ?", pos)).Error
	} else {
		return err
	}
}

func MoveDown(db *gorm.DB, value positionInterface, pos int) error {
	return MoveUp(db, value, -pos)
}

func MoveTo(db *gorm.DB, value positionInterface, pos int) error {
	return MoveUp(db, value, value.GetPosition()-pos)
}
