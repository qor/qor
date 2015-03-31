package transition

type Transition struct {
	State           string
	StateChangeLogs []StateChangeLog `sql:"-"`
}

func (transition *Transition) SetState(name string) {
	transition.State = name
}

func (transition Transition) GetState() string {
	return transition.State
}
