package sorting

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/publish"
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

func move(db *gorm.DB, value sortingInterface, pos int) (err error) {
	var startedTransaction bool
	var tx = db.Set("publish:publish_event", true)
	if t := tx.Begin(); t.Error == nil {
		startedTransaction = true
		tx = t
	}

	scope := db.NewScope(value)
	for _, field := range scope.PrimaryFields() {
		if field.DBName != "id" {
			tx = tx.Where(fmt.Sprintf("%s = ?", field.DBName), field.Field.Interface())
		}
	}

	currentPos := value.GetPosition()

	var results *gorm.DB
	if pos > 0 {
		results = tx.Model(newModel(value)).
			Where("position > ? AND position <= ?", currentPos, currentPos+pos).
			UpdateColumn("position", gorm.Expr("position - ?", 1))
	} else {
		results = tx.Model(newModel(value)).
			Where("position < ? AND position >= ?", currentPos, currentPos+pos).
			UpdateColumn("position", gorm.Expr("position + ?", 1))
	}

	if err = results.Error; err == nil {
		var rowsAffected = int(results.RowsAffected)
		if pos < 0 {
			rowsAffected = -rowsAffected
		}
		value.SetPosition(currentPos + rowsAffected)
		err = tx.Model(value).UpdateColumn("position", gorm.Expr("position + ?", rowsAffected)).Error
	}

	// Create Publish Event
	createPublishEvent(tx, value)

	if startedTransaction {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}
	return err
}

func createPublishEvent(db *gorm.DB, value interface{}) (err error) {
	// Create Publish Event in Draft Mode
	if publish.IsDraftMode(db) && publish.IsPublishableModel(value) {
		scope := db.NewScope(value)
		var sortingPublishEvent = changedSortingPublishEvent{
			Table: scope.TableName(),
		}
		for _, field := range scope.PrimaryFields() {
			sortingPublishEvent.PrimaryKeys = append(sortingPublishEvent.PrimaryKeys, field.DBName)
		}

		var result []byte
		if result, err = json.Marshal(sortingPublishEvent); err == nil {
			err = db.New().Where("publish_status = ?", publish.DIRTY).Where(map[string]interface{}{
				"name":     "changed_sorting",
				"argument": string(result),
			}).Attrs(map[string]interface{}{
				"publish_status": publish.DIRTY,
				"description":    "Changed sort order for " + scope.GetModelStruct().ModelType.Name(),
			}).FirstOrCreate(&publish.PublishEvent{}).Error
		}
	}
	return
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
