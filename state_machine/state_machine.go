package state_machine

import (
	"errors"
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
	GetState() string
}

type State struct {
	State           string
	StateChangeLogs []StateChangeLog
}

func (state *State) SetState(name string) {
	state.State = name
}

func (state *State) GetState() string {
	return state.State
}

func New(value interface{}) *StateMachine {
	return &StateMachine{states: map[string]*Event{}}
}

type StateMachine struct {
	states map[string]*Event
}

type Handle func(value interface{}, tx *gorm.DB) error

type Event struct {
	Name    string
	befores []Handle
	afters  []Handle
	enters  []Handle
	exits   []Handle
}

func (sm *StateMachine) New(name string) *Event {
	state := &Event{Name: name}
	sm.states[name] = state
	return state
}

func (sm *StateMachine) To(name string, value Stater, tx *gorm.DB) error {
	if state := sm.states[name]; state != nil {
		newTx := tx.New()
		scope := &gorm.Scope{Value: value}
		for _, before := range state.befores {
			if err := before(value, newTx); err != nil {
				return err
			}
		}

		oldState := value.GetState()
		if oldState := sm.states[oldState]; oldState != nil {
			for _, exit := range state.exits {
				if err := exit(value, newTx); err != nil {
					return err
				}
			}
		}

		value.SetState(name)

		for _, enter := range state.enters {
			if err := enter(value, newTx); err != nil {
				return err
			}
		}

		for _, after := range state.afters {
			if err := after(value, newTx); err != nil {
				return err
			}
		}

		tableName := scope.TableName()
		primaryKey := fmt.Sprintf("%v", scope.PrimaryKeyValue())
		log := StateChangeLog{ReferTable: tableName, ReferId: primaryKey, State: name}
		return newTx.Save(&log).Error
	} else {
		return errors.New("state not found")
	}
}

func (sm *Event) Before(fc Handle) *Event {
	sm.befores = append(sm.befores, fc)
	return sm
}

func (sm *Event) After(fc Handle) *Event {
	sm.afters = append(sm.afters, fc)
	return sm
}

func (sm *Event) Enter(fc Handle) *Event {
	sm.enters = append(sm.enters, fc)
	return sm
}

func (sm *Event) Exit(fc Handle) *Event {
	sm.exits = append(sm.exits, fc)
	return sm
}
