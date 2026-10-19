package nfa

type State struct {
	ID          int
	Transitions map[int32][]*State
	Epsilon     []*State
	IsAccept    bool
	Token       string
	Priority    int
}

type NFA struct {
	Start  *State
	Accept *State
}

var stateCounter int

func NewState() *State {
	state := &State{
		ID:          stateCounter,
		Transitions: make(map[int32][]*State),
		Epsilon:     make([]*State, 0),
		IsAccept:    false,
		Token:       "",
		Priority:    0,
	}

	stateCounter++

	return state
}

func (s *State) AddTransition(ch int32, target *State) {
	s.Transitions[ch] = append(s.Transitions[ch], target)
}

func (s *State) AddEpsilon(target *State) {
	s.Epsilon = append(s.Epsilon, target)
}