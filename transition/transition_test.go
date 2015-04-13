package transition_test

import (
	"errors"
	"strconv"
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
	OrderStateDraft              = "draft"
	OrderStatePaying             = "paying"
	OrderStatePaid               = "paid"
	OrderStateProcessed          = "processed"
	OrderStateDelivered          = "delivered"
	OrderStateCancelled          = "cancelled"
	OrderStateCancelledAfterPaid = "cancelled after paid"

	OrderEventCheckout = "checkout"
	OrderEventCancel   = "cancel"
)

func init() {
	ResetDb()

	OrderStateMachine.Initial(OrderStateDraft)

	OrderStateMachine.Event(OrderEventCheckout).To(OrderStatePaying).From(OrderStateDraft)
}

func CreateOrderAndExecuteTransition(order *Order, event string, t *testing.T, raiseTriggerError bool) {
	if err := testdb.Save(order).Error; err != nil {
		t.Errorf(err.Error())
	}

	if err := OrderStateMachine.Trigger(event, order, testdb); err != nil && raiseTriggerError {
		t.Errorf(err.Error())
	}
}

func TestStateTransition(t *testing.T) {
	defer ResetDb()

	order := &Order{}
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, true)

	if order.State != OrderStatePaying {
		t.Errorf("state doesn't transfered successfully")
	}
}

func TestStateChangeLog(t *testing.T) {
	defer ResetDb()

	order := &Order{}
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, true)

	var stateChangeLog transition.StateChangeLog
	testdb.First(&stateChangeLog)

	if stateChangeLog.ReferTable != "orders" {
		t.Errorf("refer table not set correctly")
	}

	if stateChangeLog.ReferId != strconv.Itoa(order.Id) {
		t.Errorf("refer id not set correctly")
	}

	if stateChangeLog.From != OrderStateDraft {
		t.Errorf("state from not set")
	}

	if stateChangeLog.To != OrderStatePaying {
		t.Errorf("state to not set")
	}
}

func TestMultipleTransitionWithOneEvent(t *testing.T) {
	defer ResetDb()

	cancellEvent := OrderStateMachine.Event(OrderEventCancel)
	cancellEvent.To(OrderStateCancelled).From(OrderStateDraft, OrderStatePaying)
	cancellEvent.To(OrderStateCancelledAfterPaid).From(OrderStatePaid, OrderStateProcessed)

	unpaidOrder := &Order{}
	unpaidOrder.State = OrderStateDraft
	CreateOrderAndExecuteTransition(unpaidOrder, OrderEventCancel, t, true)

	if unpaidOrder.State != OrderStateCancelled {
		t.Errorf("order status doesn't transitioned correctly")
	}

	paidOrder := &Order{}
	paidOrder.State = OrderStatePaid
	CreateOrderAndExecuteTransition(paidOrder, OrderEventCancel, t, true)

	if paidOrder.State != OrderStateCancelledAfterPaid {
		t.Errorf("order status doesn't transitioned correctly")
	}
}

func TestStateEnterCallback(t *testing.T) {
	defer ResetDb()

	addressAfterCheckout := "I'm an address should be set after checkout"
	OrderStateMachine.State(OrderStatePaying).Enter(func(order interface{}, tx *gorm.DB) (err error) {
		order.(*Order).Address = addressAfterCheckout
		return
	})

	order := &Order{}
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, true)

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

	order := &Order{}
	order.State = OrderStateDraft
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, true)

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

	order := &Order{}
	order.State = OrderStateDraft
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, true)

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

	order := &Order{}
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, true)

	if order.Address != addressAfterCheckout {
		t.Errorf("After callback not triggered")
	}
}

func TestRollbackTransitionOnEnterCallbackError(t *testing.T) {
	defer ResetDb()

	OrderStateMachine.State(OrderStatePaying).Enter(func(order interface{}, tx *gorm.DB) (err error) {
		order.(*Order).Address = "an address"
		return errors.New("intentional error")
	})

	order := &Order{}
	order.State = OrderStateDraft
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, false)

	testdb.First(&order, order.Id)
	if order.State != OrderStateDraft {
		t.Errorf("state transitioned on Enter callback error")
	}

	if order.Address != "" {
		t.Errorf("attribute changed on Enter callback error")
	}
}

func TestRollbackTransitionOnExitCallbackError(t *testing.T) {
	defer ResetDb()

	OrderStateMachine.State(OrderStateDraft).Exit(func(order interface{}, tx *gorm.DB) (err error) {
		order.(*Order).Address = "an address"
		return errors.New("intentional error")
	})

	order := &Order{}
	order.State = OrderStateDraft
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, false)

	testdb.First(&order, order.Id)
	if order.State != OrderStateDraft {
		t.Errorf("state transitioned on Exit callback error")
	}

	if order.Address != "" {
		t.Errorf("attribute changed on Exit callback error")
	}
}

func TestRollbackTransitionOnBeforeCallbackError(t *testing.T) {
	defer ResetDb()

	OrderStateMachine.Event(OrderEventCheckout).To(OrderStatePaying).From(OrderStateDraft).Before(func(order interface{}, tx *gorm.DB) (err error) {
		order.(*Order).Address = "an address"
		return errors.New("intentional error")
	})

	order := &Order{}
	order.State = OrderStateDraft
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, false)

	testdb.First(&order, order.Id)
	if order.State != OrderStateDraft {
		t.Errorf("state transitioned on Before callback error")
	}

	if order.Address != "" {
		t.Errorf("attribute changed on Before callback error")
	}
}

func TestRollbackTransitionOnAfterCallbackError(t *testing.T) {
	defer ResetDb()

	OrderStateMachine.Event(OrderEventCheckout).To(OrderStatePaying).From(OrderStateDraft).After(func(order interface{}, tx *gorm.DB) (err error) {
		order.(*Order).Address = "an address"
		return errors.New("intentional error")
	})

	order := &Order{}
	order.State = OrderStateDraft
	CreateOrderAndExecuteTransition(order, OrderEventCheckout, t, false)

	testdb.First(&order, order.Id)
	if order.State != OrderStateDraft {
		t.Errorf("state transitioned on Before callback error")
	}

	if order.Address != "" {
		t.Errorf("attribute changed on Before callback error")
	}
}
