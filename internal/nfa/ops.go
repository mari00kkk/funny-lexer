package nfa

func EpsilonClosure(states map[*State]bool) map[*State]bool {
	closure := make(map[*State]bool)
	stack := make([]*State, 0)

	for state := range states {
		closure[state] = true
		stack = append(stack, state)
	}

	for len(stack) > 0 {
		last := len(stack) - 1
		state := stack[last]
		stack = stack[:last]

		for _, next := range state.Epsilon {
			if !closure[next] {
				closure[next] = true
				stack = append(stack, next)
			}
		}
	}

	return closure
}

func Move(states map[*State]bool, ch int32) map[*State]bool {
	result := make(map[*State]bool)

	for state := range states {
		nextStates := state.Transitions[ch]
		for _, next := range nextStates {
			result[next] = true
		}
	}

	return result
}