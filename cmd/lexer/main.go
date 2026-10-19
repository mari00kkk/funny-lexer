package main

import (
	"fmt"
	"log"

	"funny-lexer/internal/dfa"
	"funny-lexer/internal/runner"
	"funny-lexer/internal/spec"
)

func main() {

	// Построение НКА
	funnyNFA := spec.BuildFunnyNFA()
	alphabet := spec.GetASCIIAlphabet()

	// Конструкция подмножеств (НКА в ДКА)
	rawDFA := dfa.Build(funnyNFA, alphabet)
	fmt.Printf(" Построен исходный ДКА: %d состояний\n", len(rawDFA.States))

	minDFA := dfa.Minimize(rawDFA, alphabet)
	fmt.Printf(" Минимизированный ДКА:  %d состояний\n", len(minDFA.States))

	outputPath := "transitions.json"
	err := runner.ExportTableJSON(minDFA, outputPath)
	if err != nil {
		log.Fatalf("Ошибка экспорта таблицы: %v", err)
	}
	fmt.Printf("Таблица переходов сохранена в: %s\n\n", outputPath)

	demoInput := `positive(x: int) => x > 0`
	fmt.Println("Демонстрация работы лексера")
	fmt.Printf("Входной текст:\n%s\n\nТокены:\n", demoInput)

	lexer := runner.NewLexer(minDFA)
	tokens := lexer.Tokenize(demoInput)

	for _, tok := range tokens {

		val := tok.Value
		if val == "\n" {
			val = "\\n"
		} else if val == "\t" {
			val = "\\t"
		}
		fmt.Printf("  <Type: %-10s | Value: %q>\n", tok.Type, val)
	}
}
