package runner

import (
	"funny-lexer/internal/dfa"
)

type Position struct {
	Offset int
	Line   int
	Column int
}

type Token struct {
	Type     string
	Value    string
	Position Position
}

type Lexer struct {
	dfa *dfa.DFA
}

func NewLexer(d *dfa.DFA) *Lexer {
	return &Lexer{dfa: d}
}

func (l *Lexer) Tokenize(input string) []Token {
	var tokens []Token

	runes := []rune(input)
	length := len(runes)

	i := 0

	line := 1
	column := 1

	for i < length {
		currentState := l.dfa.Start
		var lastAcceptState *dfa.State
		lastAcceptPos := -1

		j := i

		for j < length {
			ch := runes[j]

			if ch < 0 || ch > 127 {
				break
			}

			nextState, exists := currentState.Transitions[ch]
			if !exists || nextState == nil {
				break
			}

			currentState = nextState

			if currentState.IsAccept {
				lastAcceptState = currentState
				lastAcceptPos = j
			}

			j++
		}

		if lastAcceptState != nil {
			tokenValue := string(runes[i : lastAcceptPos+1])

			position := Position{
				Offset: i,
				Line:   line,
				Column: column,
			}

			tokenType := lastAcceptState.Token

			if tokenType != "WS" && tokenType != "COMMENT" {
				tokens = append(tokens, Token{
					Type:     tokenType,
					Value:    tokenValue,
					Position: position,
				})
			}

			for k := i; k <= lastAcceptPos; k++ {
				if runes[k] == '\n' {
					line++
					column = 1
				} else {
					column++
				}
			}

			i = lastAcceptPos + 1
		} else {
			tokens = append(tokens, Token{
				Type:  "ERROR",
				Value: string(runes[i]),
				Position: Position{
					Offset: i,
					Line:   line,
					Column: column,
				},
			})

			if runes[i] == '\n' {
				line++
				column = 1
			} else {
				column++
			}

			i++
		}
	}

	tokens = append(tokens, Token{
		Type: "EOF",
		Position: Position{
			Offset: i,
			Line:   line,
			Column: column,
		},
	})

	return tokens
}
