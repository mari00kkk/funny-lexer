package main

import (
	"fmt"
	"log"
	"os"

	"funny-lexer/internal/dfa"
	"funny-lexer/internal/parser"
	"funny-lexer/internal/runner"
	"funny-lexer/internal/semantic"
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

	// Построение минимизированного ДКА.
	funnyNFA := spec.BuildFunnyNFA()
	alphabet := spec.GetASCIIAlphabet()
	rawDFA := dfa.Build(funnyNFA, alphabet)
	minDFA := dfa.Minimize(rawDFA, alphabet)

	// Лексический анализ.
	lexer := runner.NewLexer(minDFA)
	tokens := lexer.Tokenize(string(source))

	// Синтаксический анализ.
	p := parser.New(tokens)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		for _, parseErr := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%d:%d: %s\n",
				parseErr.Position.Line,
				parseErr.Position.Column,
				parseErr.Message,
			)
		}
		os.Exit(1)
	}

	// Семантический анализ.
	result := semantic.Analyze(&program)

	// Печатаем единый JSON-документ: AST с типовыми аннотациями + ошибки.
	output, err := semantic.MarshalAnnotatedJSON(&program, result)
	if err != nil {
		log.Fatalf("failed to serialize annotated AST: %v", err)
	}
	fmt.Println(string(output))

	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}
