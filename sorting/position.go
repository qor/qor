package sorting

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

type positionInterface interface {
	GetPosition() int
	SetPosition(int)
}

type Position struct {
	Position int
}

func (position Position) GetPosition() int {
	return position.Position
}

func (position Position) SetPosition(pos int) {
	position.Position = pos
}

func newModel(value interface{}) interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(value)).Type()).Interface()
}

func MoveUp(db *gorm.DB, value positionInterface, pos int) error {
	clone := db
	for _, field := range db.NewScope(value).PrimaryFields() {
		primaryKey := field.Field.Interface()
		if field.DBName == "id" {
			clone = clone.Where(fmt.Sprintf("%s > ? AND %s <= (? + ?)", field.DBName, field.DBName), primaryKey, primaryKey, pos)
		} else {
			clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), primaryKey)
		}
	}

	if clone.Model(newModel(value)).Update("position", gorm.Expr("position - 1")).Error == nil {
		return db.Model(value).Update("position", gorm.Expr(fmt.Sprintf("position + %v", pos))).Error
	}
	return errors.New("failed to update position")
}

func MoveDown(db *gorm.DB, value positionInterface, pos int) error {
	clone := db
	for _, field := range db.NewScope(value).PrimaryFields() {
		primaryKey := field.Field.Interface()
		if field.DBName == "id" {
			clone = clone.Where(fmt.Sprintf("%s < ? AND %s >= (? - ?)", field.DBName, field.DBName), primaryKey, primaryKey, pos)
		} else {
			clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), primaryKey)
		}
	}

	if clone.Model(newModel(value)).Update("position", gorm.Expr("position + 1")).Error == nil {
		return db.Model(value).Update("position", gorm.Expr(fmt.Sprintf("position - %v", pos))).Error
	}
	return errors.New("failed to update position")
}

func MoveTo(db *gorm.DB, value positionInterface, pos int) error {
	clone := db
	if curPos := value.GetPosition(); pos < curPos {
		for _, field := range db.NewScope(value).PrimaryFields() {
			primaryKey := field.Field.Interface()
			if field.DBName == "id" {
				clone = clone.Where(fmt.Sprintf("%s < ? AND %s >= ?", field.DBName, field.DBName), primaryKey, pos)
			} else {
				clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), primaryKey)
			}
		}

		if clone.Model(newModel(value)).Update("position", gorm.Expr("position + 1")).Error == nil {
			return db.Model(value).Update("position", pos).Error
		}
		return errors.New("failed to update position")
	} else if pos > curPos {
		for _, field := range db.NewScope(value).PrimaryFields() {
			primaryKey := field.Field.Interface()
			if field.DBName == "id" {
				clone = clone.Where(fmt.Sprintf("%s > ? AND %s <= ?", field.DBName, field.DBName), primaryKey, pos)
			} else {
				clone = clone.Where(fmt.Sprintf("%s = ?", field.DBName), primaryKey)
			}
		}

		if clone.Model(newModel(value)).Update("position", gorm.Expr("position - 1")).Error == nil {
			return db.Model(value).Update("position", pos).Error
		}
		return errors.New("failed to update position")
	}

	return nil
}
