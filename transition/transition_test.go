package transition_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/transition"
)

type Order struct {
	Id      int
	Address string

	transition.Transition
}

var (
	testdb = func() *gorm.DB {
		db, err := gorm.Open("sqlite3", "/tmp/qor_transition_test.db")
		if err != nil {
			panic(err)
		}

		return &db
	}()

	tables []interface{}

	OrderStateMachine = transition.New(&Order{})
)

func getTables() {
	tables = []interface{}{
		&Order{},
	}

}

func ResetDb() {
	getTables()

	for _, table := range tables {
		if err := testdb.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}

		if err := testdb.AutoMigrate(table).Error; err != nil {
			panic(err)
		}
	}
}

const (
	OrderStateDraft  = "draft"
	OrderStatePaying = "paying"

	OrderEventCheckout = "checkout"
)

func init() {
	ResetDb()

	OrderStateMachine.Initialize(OrderStateDraft)

	OrderStateMachine.Event(OrderEventCheckout).To(OrderStatePaying).From(OrderStateDraft)
}

func TestStateTransition(t *testing.T) {
	order := Order{Address: "test"}
	if err := testdb.Save(&order).Error; err != nil {
		t.Errorf(err.Error())
	}

	if err := OrderStateMachine.To(OrderStatePaying, order, testdb); err != nil {
		t.Errorf(err.Error())
	}

	if order.State != OrderStatePaying {
		t.Errorf("state doesn't transfered successfully")
	}
}
