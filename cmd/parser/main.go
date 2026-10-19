package main

import (
	"fmt"
	"log"
	"os"

	"funny-lexer/internal/ast"
	"funny-lexer/internal/dfa"
	"funny-lexer/internal/parser"
	"funny-lexer/internal/runner"
	"funny-lexer/internal/spec"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <source-file>", os.Args[0])
	}

	sourcePath := os.Args[1]

	source, err := os.ReadFile(sourcePath)
	if err != nil {
		log.Fatalf("failed to read source file: %v", err)
	}

	// Строим минимизированный ДКА так же
	funnyNFA := spec.BuildFunnyNFA()
	alphabet := spec.GetASCIIAlphabet()

	rawDFA := dfa.Build(funnyNFA, alphabet)
	minDFA := dfa.Minimize(rawDFA, alphabet)

	// Лексический анализ
	lexer := runner.NewLexer(minDFA)
	tokens := lexer.Tokenize(string(source))

	// Синтаксический анализ
	p := parser.New(tokens)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		for _, parseErr := range p.Errors() {
			fmt.Printf(
				"%d:%d: %s\n",
				parseErr.Position.Line,
				parseErr.Position.Column,
				parseErr.Message,
			)
		}

		os.Exit(1)
	}

	fmt.Println("Parsing successful")
	fmt.Println()
	fmt.Print(ast.PrintProgram(program))
}
