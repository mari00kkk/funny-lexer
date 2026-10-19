package tests

import (
	"testing"

	"funny-lexer/internal/dfa"
	"funny-lexer/internal/nfa"
)

func TestStateKeyIndependentOfMapOrder(t *testing.T) {
	q1 := nfa.NewState()
	q2 := nfa.NewState()
	q3 := nfa.NewState()

	first := map[*nfa.State]bool{
		q1: true,
		q2: true,
		q3: true,
	}

	second := map[*nfa.State]bool{
		q3: true,
		q2: true,
		q1: true,
	}

	key1 := dfa.StateKey(first)
	key2 := dfa.StateKey(second)

	if key1 != key2 {
		t.Fatalf(
			"same subset produced different keys: %q and %q",
			key1, key2,
		)
	}
}
