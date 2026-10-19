package dfa

func Minimize(dfa *DFA, alphabet []int32) *DFA {
	if len(dfa.States) == 0 {
		return dfa
	}

	groups := make(map[string][]*State)
	for _, s := range dfa.States {
		key := "non_accept"
		if s.IsAccept {
			key = s.Token
		}
		groups[key] = append(groups[key], s)
	}

	var partition [][]*State
	for _, group := range groups {
		partition = append(partition, group)
	}

	worklist := make([][]*State, len(partition))
	copy(worklist, partition)

	for len(worklist) > 0 {
		A := worklist[0]
		worklist = worklist[1:]

		for _, ch := range alphabet {

			X := make(map[int]bool)
			for _, s := range dfa.States {
				if next, exists := s.Transitions[ch]; exists && next != nil {
					for _, inA := range A {
						if next.ID == inA.ID {
							X[s.ID] = true
							break
						}
					}
				}
			}

			if len(X) == 0 {
				continue
			}

			var newPartition [][]*State
			for _, Y := range partition {
				var Y1, Y2 []*State
				for _, s := range Y {
					if X[s.ID] {
						Y1 = append(Y1, s)
					} else {
						Y2 = append(Y2, s)
					}
				}

				if len(Y1) > 0 && len(Y2) > 0 {
					newPartition = append(newPartition, Y1, Y2)

					inWorklistIdx := findGroupIndex(worklist, Y)
					if inWorklistIdx != -1 {
						worklist = append(worklist[:inWorklistIdx], worklist[inWorklistIdx+1:]...)
						worklist = append(worklist, Y1, Y2)
					} else {
						if len(Y1) <= len(Y2) {
							worklist = append(worklist, Y1)
						} else {
							worklist = append(worklist, Y2)
						}
					}
				} else {
					newPartition = append(newPartition, Y)
				}
			}
			partition = newPartition
		}
	}

	return rebuildDFA(dfa, partition, alphabet)
}

func findGroupIndex(worklist [][]*State, group []*State) int {
	for i, g := range worklist {
		if sameGroup(g, group) {
			return i
		}
	}
	return -1
}

func sameGroup(g1, g2 []*State) bool {
	if len(g1) != len(g2) {
		return false
	}
	m := make(map[int]bool)
	for _, s := range g1 {
		m[s.ID] = true
	}
	for _, s := range g2 {
		if !m[s.ID] {
			return false
		}
	}
	return true
}

func rebuildDFA(orig *DFA, partition [][]*State, alphabet []int32) *DFA {
	minDFA := &DFA{States: make([]*State, 0)}
	stateToNewState := make(map[int]*State)

	for i, group := range partition {
		rep := group[0]
		newState := &State{
			ID:          i,
			IsAccept:    rep.IsAccept,
			Token:       rep.Token,
			Priority:    rep.Priority,
			Transitions: make(map[int32]*State),
		}
		minDFA.States = append(minDFA.States, newState)

		for _, s := range group {
			stateToNewState[s.ID] = newState
			if orig.Start != nil && s.ID == orig.Start.ID {
				minDFA.Start = newState
			}
		}
	}

	for _, group := range partition {
		rep := group[0]
		newState := stateToNewState[rep.ID]
		for _, ch := range alphabet {
			if target, exists := rep.Transitions[ch]; exists && target != nil {
				newState.Transitions[ch] = stateToNewState[target.ID]
			}
		}
	}

	return minDFA
}