package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFSMTransitionSuccess tests that a valid transition
// is successfully executed
func TestFSMTransitionSuccess(t *testing.T) {
	transitions := []Transition{
		{From: "closed", Event: "success", To: "closed"},
		{From: "closed", Event: "threshold_hit", To: "open"},
	}

	fsm := NewFSM("closed", transitions)
	assert.Equal(t, "closed", fsm.Current())
	fsm.Transition("threshold_hit")
	assert.Equal(t, "open", fsm.Current())
}

// TestFSMInvalidTransition tests that an invalid transition
// is rejected by the FSM
func TestFSMInvalidTransition(t *testing.T) {
	fsm := NewFSM("open", []Transition{})
	_, err := fsm.Transition("success")
	assert.Error(t, err)
}
