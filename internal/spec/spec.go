package spec

import "funny-lexer/internal/nfa"

func Range(from, to int32) nfa.NFA {
	var result nfa.NFA
	first := true
	for ch := from; ch <= to; ch++ {
		lit := nfa.Literal(ch)
		if first {
			result = lit
			first = false
		} else {
			result = nfa.Union(result, lit)
		}
	}
	return result
}

// Функции-генераторы создают НОВЫЕ состояния при каждом вызове
func Digits() nfa.NFA              { return Range('0', '9') }
func NonZeroDigits() nfa.NFA       { return Range('1', '9') }
func LettersLower() nfa.NFA        { return Range('a', 'z') }
func LettersUpper() nfa.NFA        { return Range('A', 'Z') }
func Letters() nfa.NFA             { return nfa.Union(LettersLower(), LettersUpper()) }
func LettersOrUnderscore() nfa.NFA { return nfa.Union(Letters(), nfa.Literal('_')) }

func StringLiteral(str string) nfa.NFA {
	runes := []rune(str)
	if len(runes) == 0 {
		return nfa.NFA{}
	}
	result := nfa.Literal(runes[0])
	for i := 1; i < len(runes); i++ {
		result = nfa.Concat(result, nfa.Literal(runes[i]))
	}
	return result
}

func Plus(a nfa.NFA) nfa.NFA {
	return nfa.Concat(a, nfa.StarKleene(a))
}

func AnyExceptNL() nfa.NFA {
	return nfa.Union(Range(0, 9), Range(11, 127))
}

func BuildFunnyNFA() nfa.NFA {
	superStart := nfa.NewState()

	// 1. WS: [ \t\r\n]+
	wsChar := nfa.Union(nfa.Literal(' '), nfa.Union(nfa.Literal('\t'), nfa.Union(nfa.Literal('\r'), nfa.Literal('\n'))))
	wsNFA := Plus(wsChar)
	addRule(superStart, wsNFA, "WS", 1)

	// 2. COMMENT: //[^\n]*
	commentNFA := nfa.Concat(StringLiteral("//"), nfa.StarKleene(AnyExceptNL()))
	addRule(superStart, commentNFA, "COMMENT", 1)

	// 3. INT: 0 | [1-9][0-9]*
	intZero := nfa.Literal('0')
	intNonZero := nfa.Concat(NonZeroDigits(), nfa.StarKleene(Digits()))
	intNFA := nfa.Union(intZero, intNonZero)
	addRule(superStart, intNFA, "INT", 2)

	// 4. IDENT: [a-zA-Z_][a-zA-Z0-9_]* (Изолированные свежие графы)
	identBody := nfa.Union(LettersOrUnderscore(), Digits())
	identNFA := nfa.Concat(LettersOrUnderscore(), nfa.StarKleene(identBody))
	addRule(superStart, identNFA, "IDENT", 3)

	// 5. KEYWORDS (Приоритет 4)
	keywords := []string{
		"function",
		"returns",
		"requires",
		"ensures",
		"uses",

		"if",
		"else",
		"while",
		"invariant",

		"assert",
		"assume",

		"true",
		"false",

		"and",
		"or",
		"not",

		"forall",
		"exists",

		"length",
	}

	for _, kw := range keywords {
		addRule(superStart, StringLiteral(kw), "KEYWORD", 4)
	}

	// 6. OPERATORS & DELIMITERS (Приоритеты 5 и 4)
	twoCharOps := []string{"==", "!=", "<=", ">=", "=>"}
	for _, op := range twoCharOps {
		addRule(superStart, StringLiteral(op), "OP", 5)
	}

	oneCharOps := []string{
		"+", "-", "*", "/", "<", ">", "=", ":",
		"(", ")", "[", "]", "{", "}", ",", ";",
	}
	for _, op := range oneCharOps {
		addRule(superStart, StringLiteral(op), "OP", 4)
	}

	return nfa.NFA{Start: superStart}
}

func addRule(start *nfa.State, automaton nfa.NFA, token string, priority int) {
	automaton.Accept.IsAccept = true
	automaton.Accept.Token = token
	automaton.Accept.Priority = priority
	start.AddEpsilon(automaton.Start)
}

func GetASCIIAlphabet() []int32 {
	alphabet := make([]int32, 128)
	for i := 0; i < 128; i++ {
		alphabet[i] = int32(i)
	}
	return alphabet
}
