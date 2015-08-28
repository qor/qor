package sorting

import (
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

	if startedTransaction {
		if err == nil {
			// Create Publish Event in Draft Mode
			if publish.IsDraftMode(tx) {
				tx.FirstOrCreate(map[string]interface{}{
					"Name":        "changed_sorting",
					"Argument":    scope.TableName(),
					"PublishedBy": nil,
				}).Assign(map[string]interface{}{
					"Description": "Changed sort order for " + scope.GetModelStruct().ModelType.Name(),
				}).FirstOrCreate(&publish.PublishEvent{})
			}

			tx.Commit()
		} else {
			tx.Rollback()
		}
	}

	return err
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
