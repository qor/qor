package sorting

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

type sortingInterface interface {
	GetPosition() int
	SetPosition(int)
}

type sortingDescInterface interface {
	GetPosition() int
	SetPosition(int)
	SortingDesc()
}

type Sorting struct {
	Position int `sql:"DEFAULT:NULL"`
}

func (position Sorting) GetPosition() int {
	return position.Position
}

func (position *Sorting) SetPosition(pos int) {
	position.Position = pos
}

type SortingDESC struct {
	Sorting
}

func (SortingDESC) SortingDesc() {}

func newModel(value interface{}) interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(value)).Type()).Interface()
}

func move(db *gorm.DB, value sortingInterface, pos int) error {
	clone := db
	for _, field := range db.NewScope(value).PrimaryFields() {
		if field.DBName != "id" {
			clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), field.Field.Interface())
		}
	}

	currentPos := value.GetPosition()

	if pos > 0 {
		if results := clone.Model(newModel(value)).
			Where("position > ? AND position <= ?", currentPos, currentPos+pos).
			UpdateColumn("position", gorm.Expr("position - ?", 1)); results.Error == nil {
			value.SetPosition(currentPos + pos)
			return clone.Model(value).UpdateColumn("position", gorm.Expr("position + ?", pos)).Error
		}
	} else if pos < 0 {
		if results := clone.Model(newModel(value)).
			Where("position < ? AND position >= ?", currentPos, currentPos+pos).
			UpdateColumn("position", gorm.Expr("position + ?", 1)); results.Error == nil {
			value.SetPosition(currentPos + pos)
			return clone.Model(value).UpdateColumn("position", gorm.Expr("position + ?", pos)).Error
		}
	}
	return nil
}

func MoveUp(db *gorm.DB, value sortingInterface, pos int) error {
	return move(db, value, -pos)
}

func MoveDown(db *gorm.DB, value sortingInterface, pos int) error {
	return move(db, value, pos)
}

func MoveTo(db *gorm.DB, value sortingInterface, pos int) error {
	return move(db, value, pos-value.GetPosition())
}
