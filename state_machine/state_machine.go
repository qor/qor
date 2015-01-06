package state_machine

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type StateChangeLog struct {
	Id         uint64
	ReferTable string
	ReferId    string
	State      string
	Note       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
}

type Stater interface {
	SetState(name string)
}

type State struct {
	State           string
	StateChangeLogs []StateChangeLog
}

func (state *State) SetState(name string) {
	state.State = name
}

func New(value interface{}) *StateMachine {
	return &StateMachine{}
}

type StateMachine struct {
}

type action func(value interface{}, tx *gorm.DB) error
type stateMachine struct {
	Name         string
	StateMachine *StateMachine
	befores      []action
	afters       []action
	enters       []action
	exits        []action
}

func (sm *StateMachine) New(name string) *stateMachine {
	return &stateMachine{Name: name, StateMachine: sm}
}

func (sm *StateMachine) To(name string, value Stater, tx *gorm.DB) error {
	value.SetState(name)
	scope := &gorm.Scope{Value: value}
	tableName := scope.TableName()
	primaryKey := fmt.Sprintf("%v", scope.PrimaryKeyValue())
	log := StateChangeLog{ReferTable: tableName, ReferId: primaryKey, State: name}
	tx.New().Save(&log)
	return nil
}

func (sm *stateMachine) Before(fc action) *stateMachine {
	sm.befores = append(sm.befores, fc)
	return sm
}

func (sm *stateMachine) After(fc action) *stateMachine {
	sm.afters = append(sm.afters, fc)
	return sm
}

func (sm *stateMachine) Enter(fc action) *stateMachine {
	sm.enters = append(sm.enters, fc)
	return sm
}

func (sm *stateMachine) Exit(fc action) *stateMachine {
	sm.exits = append(sm.exits, fc)
	return sm
}

// orderState.New("finish").Before().After().Do().From("ready").To("paid")
// orderState.To("finish", &order)
// order.SetState("finish")
// order.NewStateLog("finish", tableName, Id, notes)
