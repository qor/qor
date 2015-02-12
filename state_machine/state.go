package state_machine

import "strings"

type State struct {
	State           string
	StateChangeLogs []StateChangeLog `sql:"-"`
}

func (state *State) SetState(name string) {
	state.State = name
}

func (state *State) GetState() string {
	return state.State
}

func (s *State) UnmarshalJSON(data []byte) error {
	s.SetState(strings.Trim(string(data), "\""))

	return nil
}
