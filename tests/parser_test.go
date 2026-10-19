package tests

import (
	"testing"

	"funny-lexer/internal/ast"
	"funny-lexer/internal/parser"
	"funny-lexer/internal/runner"
)

func TestEmptyProgram(t *testing.T) {
	tokens := []runner.Token{
		{
			Type:     "EOF",
			Value:    "",
			Position: runner.Position{Line: 1, Column: 1},
		},
	}

	p := parser.New(tokens)
	p.Parse()

	if len(p.Errors()) == 0 {
		t.Fatal("expected error for empty program")
	}
}

func TestFormula(t *testing.T) {
	// positive(x: int) => x > 0

	tokens := []runner.Token{
		{Type: "IDENT", Value: "positive"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "IDENT", Value: "int"},
		{Type: "OP", Value: ")"},
		{Type: "OP", Value: "=>"},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ">"},
		{Type: "INT", Value: "0"},
		{Type: "EOF", Value: ""},
	}

	p := parser.New(tokens)
	program := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parser errors: %+v", p.Errors())
	}

	if len(program.Formulas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(program.Formulas))
	}

	if program.Formulas[0].Name != "positive" {
		t.Fatalf(
			"expected formula positive, got %q",
			program.Formulas[0].Name,
		)
	}

	if len(program.Functions) != 0 {
		t.Fatalf(
			"expected no functions, got %d",
			len(program.Functions),
		)
	}
}

func TestFunction(t *testing.T) {
	tokens := []runner.Token{
		{Type: "IDENT", Value: "foo"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "IDENT", Value: "int"},
		{Type: "OP", Value: ")"},
		{Type: "IDENT", Value: "returns"},
		{Type: "IDENT", Value: "y"},
		{Type: "OP", Value: ":"},
		{Type: "IDENT", Value: "int"},
		{Type: "OP", Value: "{"},

		{Type: "IDENT", Value: "y"},
		{Type: "OP", Value: "="},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "+"},
		{Type: "INT", Value: "1"},
		{Type: "OP", Value: ";"},

		{Type: "OP", Value: "}"},
		{Type: "EOF", Value: ""},
	}

	p := parser.New(tokens)
	program := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parser errors: %+v", p.Errors())
	}

	if len(program.Functions) != 1 {
		t.Fatalf(
			"expected 1 function, got %d",
			len(program.Functions),
		)
	}

	fn := program.Functions[0]

	if fn.Name != "foo" {
		t.Fatalf("expected function foo, got %q", fn.Name)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf(
			"expected 1 parameter, got %d",
			len(fn.Parameters),
		)
	}

	if len(fn.Returns) != 1 {
		t.Fatalf(
			"expected 1 return variable, got %d",
			len(fn.Returns),
		)
	}
}

func TestArrayAccess(t *testing.T) {
	// foo(a: int[]) returns x: int {
	//     x = a[0];
	// }

	tokens := []runner.Token{
		{Type: "IDENT", Value: "foo"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "a"},
		{Type: "OP", Value: ":"},
		{Type: "IDENT", Value: "int"},
		{Type: "OP", Value: "["},
		{Type: "OP", Value: "]"},
		{Type: "OP", Value: ")"},
		{Type: "IDENT", Value: "returns"},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "IDENT", Value: "int"},
		{Type: "OP", Value: "{"},

		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "="},
		{Type: "IDENT", Value: "a"},
		{Type: "OP", Value: "["},
		{Type: "INT", Value: "0"},
		{Type: "OP", Value: "]"},
		{Type: "OP", Value: ";"},

		{Type: "OP", Value: "}"},
		{Type: "EOF", Value: ""},
	}

	p := parser.New(tokens)
	program := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parser errors: %+v", p.Errors())
	}

	if len(program.Functions) != 1 {
		t.Fatalf(
			"expected 1 function, got %d",
			len(program.Functions),
		)
	}

	fn := program.Functions[0]

	block, ok := fn.Body.(*ast.Block)
	if !ok {
		t.Fatalf("expected function body to be *ast.Block, got %T", fn.Body)
	}

	if len(block.Statements) != 1 {
		t.Fatalf(
			"expected 1 statement, got %d",
			len(block.Statements),
		)
	}

	assignment, ok := block.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf(
			"expected *ast.Assignment, got %T",
			block.Statements[0],
		)
	}

	if len(assignment.Value) != 1 {
		t.Fatalf(
			"expected 1 assignment value, got %d",
			len(assignment.Value),
		)
	}

	if _, ok := assignment.Value[0].(*ast.ArrayAccessExpr); !ok {
		t.Fatalf(
			"expected *ast.ArrayAccessExpr, got %T",
			assignment.Value[0],
		)
	}
}

func TestMissingClosingBracket(t *testing.T) {
	// x = a[0;
	// Должна появиться ошибка:
	// expected "]", got ";"

	tokens := []runner.Token{
		{Type: "IDENT", Value: "foo"},
		{Type: "OP", Value: "("},
		{Type: "OP", Value: ")"},
		{Type: "IDENT", Value: "returns"},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "IDENT", Value: "int"},
		{Type: "OP", Value: "{"},

		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "="},
		{Type: "IDENT", Value: "a"},
		{Type: "OP", Value: "["},
		{Type: "INT", Value: "0"},
		{Type: "OP", Value: ";"},
		{Type: "OP", Value: "}"},
		{Type: "EOF", Value: ""},
	}

	p := parser.New(tokens)
	p.Parse()

	if len(p.Errors()) == 0 {
		t.Fatal("expected parser error")
	}

	found := false

	for _, err := range p.Errors() {
		if err.Message == `expected "]", got ";"` {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"expected missing bracket error, got %+v",
			p.Errors(),
		)
	}
}
