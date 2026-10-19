package main

import (
	"fmt"

	"funny-lexer/internal/dfa"
	"funny-lexer/internal/runner"
	"funny-lexer/internal/spec"
)

func main() {
	nfa := spec.BuildFunnyNFA()
	alphabet := spec.GetASCIIAlphabet()
	rawDFA := dfa.Build(nfa, alphabet)
	minDFA := dfa.Minimize(rawDFA, alphabet)
	lexer := runner.NewLexer(minDFA)

	for _, input := range []string{"", "0", "_my_var1", "== != <= >="} {
		fmt.Printf("Input: %q\n", input)
		fmt.Printf("Tokens: %+v\n\n", lexer.Tokenize(input))
	}
}
