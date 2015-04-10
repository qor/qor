package transition_test

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

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
		&transition.StateChangeLog{},
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
	defer ResetDb()

	order := Order{}
	if err := testdb.Save(&order).Error; err != nil {
		t.Errorf(err.Error())
	}

	if err := OrderStateMachine.To(OrderEventCheckout, &order, testdb); err != nil {
		t.Errorf(err.Error())
	}

	if order.State != OrderStatePaying {
		t.Errorf("state doesn't transfered successfully")
	}
}

func TestStateEnterCallback(t *testing.T) {
	defer ResetDb()

	addressAfterCheckout := "I'm an address should be set after checkout"
	OrderStateMachine.State(OrderStatePaying).Enter(func(order interface{}, tx *gorm.DB) (err error) {
		order.(*Order).Address = addressAfterCheckout
		return
	})

	order := Order{}
	if err := testdb.Save(&order).Error; err != nil {
		t.Errorf(err.Error())
	}

	if err := OrderStateMachine.To(OrderEventCheckout, &order, testdb); err != nil {
		t.Errorf(err.Error())
	}

	if order.Address != addressAfterCheckout {
		t.Errorf("enter callback not triggered")
	}
}

func TestStateExitCallback(t *testing.T) {
	defer ResetDb()

	var prevState string
	OrderStateMachine.State(OrderStateDraft).Exit(func(order interface{}, tx *gorm.DB) (err error) {
		prevState = order.(*Order).State
		return
	})

	order := Order{}
	order.State = OrderStateDraft
	if err := testdb.Save(&order).Error; err != nil {
		t.Errorf(err.Error())
	}

	if err := OrderStateMachine.To(OrderEventCheckout, &order, testdb); err != nil {
		t.Errorf(err.Error())
	}

	if prevState != OrderStateDraft {
		t.Errorf("exit callback triggered after state change")
	}
}

func TestEventBeforeCallback(t *testing.T) {
	defer ResetDb()

	var prevState string
	OrderStateMachine.Event(OrderEventCheckout).To(OrderStatePaying).From(OrderStateDraft).Before(func(order interface{}, tx *gorm.DB) (err error) {
		prevState = order.(*Order).State
		return
	})

	order := Order{}
	order.State = OrderStateDraft
	if err := testdb.Save(&order).Error; err != nil {
		t.Errorf(err.Error())
	}

	if err := OrderStateMachine.To(OrderEventCheckout, &order, testdb); err != nil {
		t.Errorf(err.Error())
	}

	if prevState != OrderStateDraft {
		t.Errorf("Before callback triggered after state change")
	}
}

func TestEventAfterCallback(t *testing.T) {
	defer ResetDb()

	addressAfterCheckout := "I'm an address should be set after checkout"
	OrderStateMachine.Event(OrderEventCheckout).To(OrderStatePaying).From(OrderStateDraft).After(func(order interface{}, tx *gorm.DB) (err error) {
		order.(*Order).Address = addressAfterCheckout
		return
	})

	order := Order{}
	if err := testdb.Save(&order).Error; err != nil {
		t.Errorf(err.Error())
	}

	if err := OrderStateMachine.To(OrderEventCheckout, &order, testdb); err != nil {
		t.Errorf(err.Error())
	}

	if order.Address != addressAfterCheckout {
		t.Errorf("After callback not triggered")
	}
}
