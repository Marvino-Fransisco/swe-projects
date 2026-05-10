package pkg

import "fmt"

type Event string

const (
	EventSuccess       Event = "success"
	EventFailure       Event = "failure"
	EventThresholdHit  Event = "threshold_hit"
	EventCooldownTimer Event = "cooldown_timer"
	EventForceOpen     Event = "force_open"
)

type Transition struct {
	From  string
	Event Event
	To    string
}

type FSM struct {
	current     string
	transitions []Transition
}

func NewFSM(initial string, transitions []Transition) *FSM {
	return &FSM{
		current:     initial,
		transitions: transitions,
	}
}

func (f *FSM) Current() string {
	return f.current
}

func (f *FSM) Transition(event Event) (string, error) {
	for _, t := range f.transitions {
		if t.From == f.current && t.Event == event {
			f.current = t.To
			return t.To, nil
		}
	}
	return f.current, fmt.Errorf("no transition from %q on event %q", f.current, event)
}

func (f *FSM) CanTransition(event Event) bool {
	for _, t := range f.transitions {
		if t.From == f.current && t.Event == event {
			return true
		}
	}
	return false
}
