package transition

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

func New(value interface{}) *StateMachine {
	return &StateMachine{events: map[string]*Event{}}
}

type StateMachine struct {
	states map[string]*State
	events map[string]*Event
}

func (sm *StateMachine) State(name string) *State {
	event := &State{Name: name}
	sm.states[name] = event
	return event
}

func (sm *StateMachine) Event(name string) *Event {
	event := &Event{Name: name}
	sm.events[name] = event
	return event
}

func (sm *StateMachine) To(name string, value Stater, tx *gorm.DB) error {
	stateWas := value.GetState()

	newTx := tx.New()

	if event := sm.events[name]; event != nil {
		// State: exit
		if state, ok := sm.states[stateWas]; ok {
			for _, exit := range state.exits {
				if err := exit(value, newTx); err != nil {
					return err
				}
			}
		}

		// Event: before
		for _, before := range event.befores {
			if err := before(value, newTx); err != nil {
				return err
			}
		}

		value.SetState(name)

		// Event: after
		for _, after := range event.afters {
			if err := after(value, newTx); err != nil {
				return err
			}
		}

		scope := newTx.NewScope(value)
		primaryKey := fmt.Sprintf("%v", scope.PrimaryKeyValue())
		log := StateChangeLog{ReferTable: scope.TableName(), ReferId: primaryKey, State: name}
		return newTx.Save(&log).Error
	}
	return errors.New("state not found")
}

type State struct {
	Name   string
	enters []func(value interface{}, tx *gorm.DB) error
	exits  []func(value interface{}, tx *gorm.DB) error
}

func (state *State) Enter(fc func(value interface{}, tx *gorm.DB) error) *State {
	state.enters = append(state.enters, fc)
	return state
}

func (state *State) Exit(fc func(value interface{}, tx *gorm.DB) error) *State {
	state.exits = append(state.exits, fc)
	return state
}

type Event struct {
	Name    string
	befores []func(value interface{}, tx *gorm.DB) error
	afters  []func(value interface{}, tx *gorm.DB) error
}

func (event *Event) Before(fc func(value interface{}, tx *gorm.DB) error) *Event {
	event.befores = append(event.befores, fc)
	return event
}

func (event *Event) After(fc func(value interface{}, tx *gorm.DB) error) *Event {
	event.afters = append(event.afters, fc)
	return event
}
