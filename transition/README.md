# Introduction

transition is a Golang state machine implementation. rely on github.com/jinzhu/gorm

# GO Compatibility ?

# Installation

    go get github.com/theplant/qor3/transition

# Usage

### 1. Add transition to object struct
    type Order struct {
        ID        uint
        ...

        // Add transition to Order
        // type Transition struct {
        //   State           string
        //   StateChangeLogs []StateChangeLog `sql:"-"`
        // }
        transition.Transition
    }

### 2. Define states and events

    var OrderStateMachine transition.New(&Order{})

    func init() {
		// Define initial state
		OrderStateMachine.Initial("draft")

        // Define what to do when enter a state and exit a state. See Callbacks for detail.
		OrderStateMachine.State("paying").Enter(func(order interface{}, tx *gorm.DB) (err error) {
            // To get order object use 'order.(*Order)'
    		// business logic here
    		return
    	}).Exit(func(order interface{}, tx *gorm.DB) (err error) {
            // business logic here
            return
        })

        // Define event and what to do before perform transition and after transition. See Callbacks for detail.
        OrderStateMachine.Event("checkout").From("draft").To("paying").Before(func(order interface{}, tx *gorm.DB) (err error) {
            // business logic here
            return
        }).After(func(order interface{}, tx *gorm.DB) (err error) {
            // business logic here
            return
        })

        // Different state transition for one event
        cancellEvent := OrderStateMachine.Event("cancel")
        cancellEvent.From("draft", "paying").To("canceled")
        cancellEvent.From("paid", "processed").To("canceled_after_paid")
    }

### 3. Transfer state

    // func (sm *StateMachine) Trigger(name string, value Stater, tx *gorm.DB) error {
    OrderStatemachine.Trigger("checkout", *order, db)

# Callbacks

## There are 2 callbacks for state.

### Enter
This will be performed when entering a state. Object is not persisted and object state is not changed.

    OrderStateMachine.State("paying").Enter(func(order interface{}, tx *gorm.DB) (err error) {
        // To get order object use 'order.(*Order)'
        // business logic here
        return
    })

### Exit
This will be performed when exiting a state. Object is not persisted and object state is not changed.

    OrderStateMachine.State("paying").Exit(func(order interface{}, tx *gorm.DB) (err error) {
        // business logic here
        return
    })

## And 2 callbacks for event.

### Before
This will be performed before the event performed. Object is not persisted and object state is not changed.

    OrderStateMachine.Event("checkout").From("draft").To("paying").Before(func(order interface{}, tx *gorm.DB) (err error) {
        // business logic here
        return
    })

### After
This will be performed after the event performed. Object is persisted.

    OrderStateMachine.Event("checkout").From("draft").To("paying").Before(func(order interface{}, tx *gorm.DB) (err error) {
        // business logic here
        return
    })

# Features

### State change log
not implemented

# Copyright
Copyright © The Plant
