package tests

import (
	"reflect"
	"testing"

	"funny-lexer/internal/dfa"
	"funny-lexer/internal/runner"
	"funny-lexer/internal/spec"
)

// newTestLexerForLexerTests строит настоящий лексер языка Funny.
func newTestLexerForLexerTests(t *testing.T) *runner.Lexer {
	t.Helper()

	nfa := spec.BuildFunnyNFA()
	alphabet := spec.GetASCIIAlphabet()

	rawDFA := dfa.Build(nfa, alphabet)
	minDFA := dfa.Minimize(rawDFA, alphabet)

	if minDFA == nil {
		t.Fatal("dfa.Minimize returned nil")
	}

	if minDFA.Start == nil {
		t.Fatal("minimized DFA has no start state")
	}

	return runner.NewLexer(minDFA)
}

// expectedToken описывает ожидаемые тип и значение токена.
type expectedToken struct {
	tokenType string
	value     string
}

func setupLexer() *runner.Lexer {
	nfa := spec.BuildFunnyNFA()
	alphabet := spec.GetASCIIAlphabet()
	rawDFA := dfa.Build(nfa, alphabet)
	minDFA := dfa.Minimize(rawDFA, alphabet)
	return runner.NewLexer(minDFA)
}

// checkTokens проверяет типы и значения всех токенов.
// Позиции проверяются отдельно в соответствующих тестах.
func checkTokens(
	t *testing.T,
	got []runner.Token,
	want []expectedToken,
) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf(
			"unexpected number of tokens: got %d, want %d\n"+
				"got:  %#v\nwant: %#v",
			len(got), len(want), got, want,
		)
	}

	for i := range want {
		if got[i].Type != want[i].tokenType {
			t.Errorf(
				"token %d: type = %q, want %q",
				i, got[i].Type, want[i].tokenType,
			)
		}

		if got[i].Value != want[i].value {
			t.Errorf(
				"token %d: value = %q, want %q",
				i, got[i].Value, want[i].value,
			)
		}
	}
}

// проверяет пустой ввод.
func TestEmptyAndWhitespace(t *testing.T) {
	lexer := setupLexer()

	tokens := lexer.Tokenize("")
	if len(tokens) != 1 || tokens[0].Type != "EOF" {
		t.Fatalf("expected only EOF for empty input, got %+v", tokens)
	}

	tokens = lexer.Tokenize("  \t\r\n")
	if len(tokens) != 1 || tokens[0].Type != "EOF" {
		t.Fatalf("expected only EOF for whitespace input, got %+v", tokens)
	}
}

// проверяет целочисленные литералы.
func TestNumbers(t *testing.T) {
	lexer := setupLexer()

	tokens := lexer.Tokenize("0")
	if len(tokens) != 2 ||
		tokens[0].Type != "INT" ||
		tokens[0].Value != "0" ||
		tokens[1].Type != "EOF" {
		t.Fatalf("expected INT('0') followed by EOF, got %+v", tokens)
	}

	tokens = lexer.Tokenize("00")
	if len(tokens) != 3 ||
		tokens[0].Value != "0" ||
		tokens[1].Value != "0" ||
		tokens[2].Type != "EOF" {
		t.Fatalf("expected two INT('0') tokens followed by EOF, got %+v", tokens)
	}

	tokens = lexer.Tokenize("01")
	if len(tokens) != 3 ||
		tokens[0].Type != "INT" || tokens[0].Value != "0" ||
		tokens[1].Type != "INT" || tokens[1].Value != "1" ||
		tokens[2].Type != "EOF" {
		t.Fatalf("expected INT('0'), INT('1'), EOF, got %+v", tokens)
	}
}

// TestLexerIdentifiers проверяет идентификаторы.
func TestLexerIdentifiers(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("x abc _value value123 A1")

	want := []expectedToken{
		{tokenType: "IDENT", value: "x"},
		{tokenType: "IDENT", value: "abc"},
		{tokenType: "IDENT", value: "_value"},
		{tokenType: "IDENT", value: "value123"},
		{tokenType: "IDENT", value: "A1"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)
}

// TestLexerKeywords проверяет ключевые слова.
func TestLexerKeywords(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize(
		"function returns while if else assert assume invariant length",
	)

	want := []expectedToken{
		{tokenType: "KEYWORD", value: "function"},
		{tokenType: "KEYWORD", value: "returns"},
		{tokenType: "KEYWORD", value: "while"},
		{tokenType: "KEYWORD", value: "if"},
		{tokenType: "KEYWORD", value: "else"},
		{tokenType: "KEYWORD", value: "assert"},
		{tokenType: "KEYWORD", value: "assume"},
		{tokenType: "KEYWORD", value: "invariant"},
		{tokenType: "KEYWORD", value: "length"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)
}

// TestLexerKeywordPrefix проверяет, что идентификатор,
// начинающийся с ключевого слова, не разбивается на части.
func TestLexerKeywordPrefix(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("if ifx functionName")

	want := []expectedToken{
		{tokenType: "KEYWORD", value: "if"},
		{tokenType: "IDENT", value: "ifx"},
		{tokenType: "IDENT", value: "functionName"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)
}

// TestLexerTwoCharacterOperators проверяет двухсимвольные операторы.
func TestLexerTwoCharacterOperators(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("== != <= >=")

	want := []expectedToken{
		{tokenType: "OP", value: "=="},
		{tokenType: "OP", value: "!="},
		{tokenType: "OP", value: "<="},
		{tokenType: "OP", value: ">="},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)
}

// TestLexerSingleCharacterOperators проверяет односимвольные
// операторы и разделители.
func TestLexerSingleCharacterOperators(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("+ - * / < > = ( ) [ ] { } , ;")

	values := []string{
		"+", "-", "*", "/", "<", ">", "=",
		"(", ")", "[", "]", "{", "}", ",", ";",
	}

	want := make([]expectedToken, 0, len(values)+1)

	for _, value := range values {
		want = append(want, expectedToken{
			tokenType: "OP",
			value:     value,
		})
	}

	want = append(want, expectedToken{
		tokenType: "EOF",
		value:     "",
	})

	checkTokens(t, got, want)
}

// TestLexerWhitespace проверяет пропуск пробелов, табуляций
// и переводов строки.
func TestLexerWhitespace(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize(" \t\r\nx \n y")

	want := []expectedToken{
		{tokenType: "IDENT", value: "x"},
		{tokenType: "IDENT", value: "y"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)
}

// TestLexerComments проверяет пропуск однострочных комментариев.
func TestLexerComments(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	input := "x // comment\ny"

	got := lexer.Tokenize(input)

	want := []expectedToken{
		{tokenType: "IDENT", value: "x"},
		{tokenType: "IDENT", value: "y"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)

	// "y" находится в начале второй строки.
	if got[1].Position.Line != 2 {
		t.Errorf(
			"y: line = %d, want 2",
			got[1].Position.Line,
		)
	}

	if got[1].Position.Column != 1 {
		t.Errorf(
			"y: column = %d, want 1",
			got[1].Position.Column,
		)
	}
}

// TestLexerPositions проверяет позиции токенов.
// Offset отсчитывается от начала входа с нуля,
// Line и Column — с единицы.
func TestLexerPositions(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("x +\ny")

	if len(got) != 4 {
		t.Fatalf("got %d tokens, want 4", len(got))
	}

	wantPositions := []runner.Position{
		{Offset: 0, Line: 1, Column: 1}, // x
		{Offset: 2, Line: 1, Column: 3}, // +
		{Offset: 4, Line: 2, Column: 1}, // y
		{Offset: 5, Line: 2, Column: 2}, // EOF
	}

	for i, want := range wantPositions {
		if got[i].Position != want {
			t.Errorf(
				"token %d (%q): position = %+v, want %+v",
				i,
				got[i].Value,
				got[i].Position,
				want,
			)
		}
	}
}

// TestLexerInvalidCharacter проверяет недопустимый ASCII-символ.
func TestLexerInvalidCharacter(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("x @ y")

	want := []expectedToken{
		{tokenType: "IDENT", value: "x"},
		{tokenType: "ERROR", value: "@"},
		{tokenType: "IDENT", value: "y"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)

	if got[1].Position.Column != 3 {
		t.Errorf(
			"error column = %d, want 3",
			got[1].Position.Column,
		)
	}
}

// TestLexerUnicodeCharacter проверяет обработку символа,
// который не входит в ASCII-алфавит автомата.
func TestLexerUnicodeCharacter(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("aéb")

	want := []expectedToken{
		{tokenType: "IDENT", value: "a"},
		{tokenType: "ERROR", value: "é"},
		{tokenType: "IDENT", value: "b"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)

	// В текущей реализации Offset измеряется в рунах,
	// а не в байтах UTF-8.
	wantPositions := []runner.Position{
		{Offset: 0, Line: 1, Column: 1},
		{Offset: 1, Line: 1, Column: 2},
		{Offset: 2, Line: 1, Column: 3},
		{Offset: 3, Line: 1, Column: 4},
	}

	for i, wantPosition := range wantPositions {
		if got[i].Position != wantPosition {
			t.Errorf(
				"token %d (%q): position = %+v, want %+v",
				i,
				got[i].Value,
				got[i].Position,
				wantPosition,
			)
		}
	}
}

// TestLexerFunnyExpression проверяет токенизацию
// небольшого выражения языка Funny.
func TestLexerFunnyExpression(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("x = 2 + 3 * 4;")

	want := []expectedToken{
		{tokenType: "IDENT", value: "x"},
		{tokenType: "OP", value: "="},
		{tokenType: "INT", value: "2"},
		{tokenType: "OP", value: "+"},
		{tokenType: "INT", value: "3"},
		{tokenType: "OP", value: "*"},
		{tokenType: "INT", value: "4"},
		{tokenType: "OP", value: ";"},
		{tokenType: "EOF", value: ""},
	}

	checkTokens(t, got, want)
}

// TestLexerNoUnexpectedEmptyTokens проверяет,
// что ни один токен, кроме EOF, не имеет пустого значения.
func TestLexerNoUnexpectedEmptyTokens(t *testing.T) {
	lexer := newTestLexerForLexerTests(t)

	got := lexer.Tokenize("x + 10")

	for i, token := range got {
		if token.Type != "EOF" && token.Value == "" {
			t.Errorf("token %d has an empty value: %+v", i, token)
		}
	}

	if len(got) == 0 || got[len(got)-1].Type != "EOF" {
		t.Fatal("token stream must end with EOF")
	}

	// Проверяем, что EOF встречается ровно один раз.
	eofCount := 0

	for _, token := range got {
		if token.Type == "EOF" {
			eofCount++
		}
	}

	if !reflect.DeepEqual(eofCount, 1) {
		t.Errorf("got %d EOF tokens, want 1", eofCount)
	}
}
