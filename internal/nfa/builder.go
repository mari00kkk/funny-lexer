package nfa

func Union(a, b NFA) NFA {
	start := NewState()
	accept := NewState()

	start.AddEpsilon(a.Start)
	start.AddEpsilon(b.Start)

	a.Accept.AddEpsilon(accept)
	b.Accept.AddEpsilon(accept)

	return NFA{ Start:  start, Accept: accept}
}

func Literal(ch int32) NFA {
	start := NewState()
	accept := NewState()

	start.AddTransition(ch, accept)

	return NFA{ Start:  start, Accept: accept}
}

func Concat(a, b NFA) NFA {
	a.Accept.AddEpsilon(b.Start)

	return NFA{ Start:  a.Start, Accept: b.Accept}
}

func StarKleene(a NFA) NFA {
	start := NewState()
	accept := NewState()

	start.AddEpsilon(a.Start)
	start.AddEpsilon(accept)

	a.Accept.AddEpsilon(a.Start)
	a.Accept.AddEpsilon(accept)

	return NFA{ Start:  start, Accept: accept}
}