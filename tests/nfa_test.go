package tests

import (
	"testing"

	"funny-lexer/internal/nfa"
)

// accepts проверяет, принимает ли НКА заданную строку.
//
// На каждом шаге:
// 1. Берём ε-замыкание начального множества состояний.
// 2. Выполняем переход по очередному символу.
// 3. Снова вычисляем ε-замыкание.
// 4. Проверяем, содержится ли конечное состояние в результате.
func accepts(n nfa.NFA, input string) bool {
	states := nfa.EpsilonClosure(map[*nfa.State]bool{
		n.Start: true,
	})

	for _, ch := range input {
		moved := nfa.Move(states, int32(ch))
		states = nfa.EpsilonClosure(moved)
	}

	return states[n.Accept]
}

// TestNewState проверяет начальное состояние нового узла.
func TestNewState(t *testing.T) {
	state := nfa.NewState()

	if state == nil {
		t.Fatal("nfa.NewState returned nil")
	}

	if state.Transitions == nil {
		t.Fatal("Transitions must be initialized")
	}

	if state.Epsilon == nil {
		t.Fatal("Epsilon must be initialized")
	}

	if state.IsAccept {
		t.Error("new state must not be accepting")
	}

	if state.Token != "" {
		t.Errorf("Token = %q, want empty string", state.Token)
	}

	if state.Priority != 0 {
		t.Errorf("Priority = %d, want 0", state.Priority)
	}
}

// TestNewStateUniqueIDs проверяет уникальность ID
// двух последовательно созданных состояний.
func TestNewStateUniqueIDs(t *testing.T) {
	first := nfa.NewState()
	second := nfa.NewState()

	if first.ID == second.ID {
		t.Errorf("states have the same ID: %d", first.ID)
	}
}

// TestAddTransition проверяет добавление перехода по символу.
func TestAddTransition(t *testing.T) {
	from := nfa.NewState()
	to := nfa.NewState()

	from.AddTransition('a', to)

	transitions := from.Transitions['a']

	if len(transitions) != 1 {
		t.Fatalf("got %d transitions, want 1", len(transitions))
	}

	if transitions[0] != to {
		t.Fatal("transition points to the wrong state")
	}
}

// TestAddTransitionMultipleTargets проверяет,
// что переходы по одному символу могут вести в несколько состояний.
func TestAddTransitionMultipleTargets(t *testing.T) {
	from := nfa.NewState()
	first := nfa.NewState()
	second := nfa.NewState()

	from.AddTransition('a', first)
	from.AddTransition('a', second)

	transitions := from.Transitions['a']

	if len(transitions) != 2 {
		t.Fatalf("got %d transitions, want 2", len(transitions))
	}

	foundFirst := false
	foundSecond := false

	for _, state := range transitions {
		if state == first {
			foundFirst = true
		}

		if state == second {
			foundSecond = true
		}
	}

	if !foundFirst || !foundSecond {
		t.Error("not all transition targets were saved")
	}
}

// TestAddEpsilon проверяет добавление ε-перехода.
func TestAddEpsilon(t *testing.T) {
	from := nfa.NewState()
	to := nfa.NewState()

	from.AddEpsilon(to)

	if len(from.Epsilon) != 1 {
		t.Fatalf("got %d epsilon transitions, want 1", len(from.Epsilon))
	}

	if from.Epsilon[0] != to {
		t.Fatal("epsilon transition points to the wrong state")
	}
}

// TestEpsilonClosure проверяет, что ε-замыкание содержит
// исходное состояние и все состояния, достижимые по ε-переходам.
func TestEpsilonClosure(t *testing.T) {
	first := nfa.NewState()
	second := nfa.NewState()
	third := nfa.NewState()
	unreachable := nfa.NewState()

	first.AddEpsilon(second)
	second.AddEpsilon(third)

	closure := nfa.EpsilonClosure(map[*nfa.State]bool{
		first: true,
	})

	for _, state := range []*nfa.State{first, second, third} {
		if !closure[state] {
			t.Errorf("state %d is missing from epsilon closure", state.ID)
		}
	}

	if closure[unreachable] {
		t.Error("unreachable state must not be in epsilon closure")
	}
}

// TestEpsilonClosureWithCycle проверяет ε-замыкание,
// когда в графе есть цикл.
func TestEpsilonClosureWithCycle(t *testing.T) {
	first := nfa.NewState()
	second := nfa.NewState()
	third := nfa.NewState()

	first.AddEpsilon(second)
	second.AddEpsilon(first)
	second.AddEpsilon(third)

	closure := nfa.EpsilonClosure(map[*nfa.State]bool{
		first: true,
	})

	for _, state := range []*nfa.State{first, second, third} {
		if !closure[state] {
			t.Errorf("state %d is missing from closure", state.ID)
		}
	}

	if len(closure) != 3 {
		t.Errorf("closure size = %d, want 3", len(closure))
	}
}

// TestEpsilonClosureMultipleStarts проверяет замыкание
// множества, содержащего несколько исходных состояний.
func TestEpsilonClosureMultipleStarts(t *testing.T) {
	first := nfa.NewState()
	second := nfa.NewState()
	third := nfa.NewState()

	first.AddEpsilon(third)
	second.AddEpsilon(third)

	closure := nfa.EpsilonClosure(map[*nfa.State]bool{
		first:  true,
		second: true,
	})

	for _, state := range []*nfa.State{first, second, third} {
		if !closure[state] {
			t.Errorf("state %d is missing from closure", state.ID)
		}
	}
}

// TestMove проверяет переход по заданному символу.
func TestMove(t *testing.T) {
	first := nfa.NewState()
	second := nfa.NewState()
	third := nfa.NewState()

	first.AddTransition('a', second)
	first.AddTransition('b', third)

	result := nfa.Move(map[*nfa.State]bool{
		first: true,
	}, 'a')

	if len(result) != 1 {
		t.Fatalf("got %d states, want 1", len(result))
	}

	if !result[second] {
		t.Error("expected destination state is missing")
	}

	if result[third] {
		t.Error("unexpected destination state")
	}
}

// TestMoveMultipleStates проверяет переходы из множества состояний.
func TestMoveMultipleStates(t *testing.T) {
	first := nfa.NewState()
	second := nfa.NewState()
	target1 := nfa.NewState()
	target2 := nfa.NewState()

	first.AddTransition('a', target1)
	second.AddTransition('a', target2)

	result := nfa.Move(map[*nfa.State]bool{
		first:  true,
		second: true,
	}, 'a')

	if len(result) != 2 {
		t.Fatalf("got %d states, want 2", len(result))
	}

	if !result[target1] || !result[target2] {
		t.Error("nfa.Move did not return all destination states")
	}
}

// TestMoveNoTransition проверяет отсутствие перехода
// по символу, для которого переход не задан.
func TestMoveNoTransition(t *testing.T) {
	state := nfa.NewState()
	target := nfa.NewState()

	state.AddTransition('a', target)

	result := nfa.Move(map[*nfa.State]bool{
		state: true,
	}, 'b')

	if len(result) != 0 {
		t.Errorf("got %d states, want an empty set", len(result))
	}
}

// TestLiteral проверяет автомат, распознающий один символ.
func TestLiteral(t *testing.T) {
	n := nfa.Literal('a')

	tests := []struct {
		input string
		want  bool
	}{
		{"a", true},
		{"", false},
		{"b", false},
		{"aa", false},
		{"ab", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := accepts(n, test.input)

			if got != test.want {
				t.Errorf(
					"accepts(%q) = %v, want %v",
					test.input, got, test.want,
				)
			}
		})
	}
}

// TestUnion проверяет объединение двух языков.
// nfa.Union(nfa.Literal('a'), nfa.Literal('b')) должен принимать a или b.
func TestUnion(t *testing.T) {
	n := nfa.Union(nfa.Literal('a'), nfa.Literal('b'))

	tests := []struct {
		input string
		want  bool
	}{
		{"a", true},
		{"b", true},
		{"", false},
		{"c", false},
		{"ab", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := accepts(n, test.input)

			if got != test.want {
				t.Errorf(
					"accepts(%q) = %v, want %v",
					test.input, got, test.want,
				)
			}
		})
	}
}

// TestConcat проверяет конкатенацию автоматов.
// nfa.Concat(nfa.Literal('a'), nfa.Literal('b')) должен принимать ab.
func TestConcat(t *testing.T) {
	n := nfa.Concat(nfa.Literal('a'), nfa.Literal('b'))

	tests := []struct {
		input string
		want  bool
	}{
		{"ab", true},
		{"", false},
		{"a", false},
		{"b", false},
		{"ba", false},
		{"abc", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := accepts(n, test.input)

			if got != test.want {
				t.Errorf(
					"accepts(%q) = %v, want %v",
					test.input, got, test.want,
				)
			}
		})
	}
}

// TestStarKleene проверяет операцию звезды Клини.
// nfa.StarKleene(nfa.Literal('a')) должен принимать ноль
// или любое количество символов a.
func TestStarKleene(t *testing.T) {
	n := nfa.StarKleene(nfa.Literal('a'))

	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"a", true},
		{"aa", true},
		{"aaa", true},
		{"b", false},
		{"ab", false},
		{"ba", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := accepts(n, test.input)

			if got != test.want {
				t.Errorf(
					"accepts(%q) = %v, want %v",
					test.input, got, test.want,
				)
			}
		})
	}
}

// TestStarKleeneOfUnion проверяет повторение языка {a, b}.
func TestStarKleeneOfUnion(t *testing.T) {
	n := nfa.StarKleene(nfa.Union(nfa.Literal('a'), nfa.Literal('b')))

	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"a", true},
		{"b", true},
		{"ab", true},
		{"ba", true},
		{"abba", true},
		{"c", false},
		{"abc", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := accepts(n, test.input)

			if got != test.want {
				t.Errorf(
					"accepts(%q) = %v, want %v",
					test.input, got, test.want,
				)
			}
		})
	}
}
