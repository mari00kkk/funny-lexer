package dfa

import (
	"fmt"
	"sort"
	"strings"

	"funny-lexer/internal/nfa"
)

type State struct {
	ID          int
	NFAStates   map[*nfa.State]bool
	IsAccept    bool
	Token       string
	Priority    int
	Transitions map[int32]*State
}

type DFA struct {
	Start  *State
	States []*State
}

func StateKey(states map[*nfa.State]bool) string {
	ptrs := make([]string, 0, len(states))
	for state := range states {
		ptrs = append(ptrs, fmt.Sprintf("%p", state))
	}
	sort.Strings(ptrs)
	return strings.Join(ptrs, ",")
}

func resolveToken(nfaStates map[*nfa.State]bool) (bool, string, int) {
	isAccept := false
	bestToken := ""
	bestPriority := -1
	firstAccept := true

	for s := range nfaStates {
		if s.IsAccept {
			isAccept = true
			if firstAccept || s.Priority > bestPriority {
				bestToken = s.Token
				bestPriority = s.Priority
				firstAccept = false
			}
		}
	}
	return isAccept, bestToken, bestPriority
}

func Build(automaton nfa.NFA, alphabet []int32) *DFA {
	dfa := &DFA{
		States: make([]*State, 0),
	}

	startSet := map[*nfa.State]bool{
		automaton.Start: true,
	}

	startSet = nfa.EpsilonClosure(startSet)
	isAccept, token, priority := resolveToken(startSet)

	startState := &State{
		ID:          0,
		NFAStates:   startSet,
		IsAccept:    isAccept,
		Token:       token,
		Priority:    priority,
		Transitions: make(map[int32]*State),
	}

	dfa.Start = startState
	dfa.States = append(dfa.States, startState)

	stateMap := make(map[string]*State)
	stateMap[StateKey(startSet)] = startState

	queue := []*State{startState}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, ch := range alphabet {
			nextSet := nfa.Move(current.NFAStates, ch)
			nextSet = nfa.EpsilonClosure(nextSet)

			if len(nextSet) == 0 {
				continue
			}

			key := StateKey(nextSet)
			nextState, exists := stateMap[key]

			if !exists {
				acc, tok, prio := resolveToken(nextSet)
				nextState = &State{
					ID:          len(dfa.States),
					NFAStates:   nextSet,
					IsAccept:    acc,
					Token:       tok,
					Priority:    prio,
					Transitions: make(map[int32]*State),
				}

				dfa.States = append(dfa.States, nextState)
				stateMap[key] = nextState
				queue = append(queue, nextState)
			}

			current.Transitions[ch] = nextState
		}
	}

	return dfa
}