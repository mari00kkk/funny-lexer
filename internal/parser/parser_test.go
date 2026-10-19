package parser

import (
	"funny-lexer/internal/ast"
	"funny-lexer/internal/runner"
	"testing"
)

func TestParseUnaryMinus(t *testing.T) {
	tokens := []runner.Token{
		{Type: "OP", Value: "-"},
		{Type: "INT", Value: "5"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	expr := p.parseExpr()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	unary, ok := expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}

	if unary.Op != "-" {
		t.Fatalf("expected unary '-', got %q", unary.Op)
	}

	number, ok := unary.X.(*ast.NumberExpr)
	if !ok {
		t.Fatalf("expected NumberExpr inside unary expression, got %T", unary.X)
	}

	if number.Value != "5" {
		t.Fatalf("expected value 5, got %q", number.Value)
	}
}

func TestParseFunctionCall(t *testing.T) {
	tokens := []runner.Token{
		{Type: "IDENT", Value: "foo"},
		{Type: "OP", Value: "("},
		{Type: "INT", Value: "1"},
		{Type: "OP", Value: ","},
		{Type: "INT", Value: "2"},
		{Type: "OP", Value: ")"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	expr := p.parseExpr()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	call, ok := expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", expr)
	}

	if call.Name != "foo" {
		t.Fatalf("expected function name foo, got %q", call.Name)
	}

	if len(call.Args) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(call.Args))
	}

	first, ok := call.Args[0].(*ast.NumberExpr)
	if !ok || first.Value != "1" {
		t.Fatalf("expected first argument to be NumberExpr(1), got %T", call.Args[0])
	}

	second, ok := call.Args[1].(*ast.NumberExpr)
	if !ok || second.Value != "2" {
		t.Fatalf("expected second argument to be NumberExpr(2), got %T", call.Args[1])
	}
}

func TestParseArrayAccess(t *testing.T) {
	tokens := []runner.Token{
		{Type: "IDENT", Value: "a"},
		{Type: "OP", Value: "["},
		{Type: "INT", Value: "2"},
		{Type: "OP", Value: "]"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	expr := p.parseExpr()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	access, ok := expr.(*ast.ArrayAccessExpr)
	if !ok {
		t.Fatalf("expected ArrayAccessExpr, got %T", expr)
	}

	if access.Array != "a" {
		t.Fatalf("expected array name a, got %q", access.Array)
	}

	index, ok := access.Index.(*ast.NumberExpr)
	if !ok {
		t.Fatalf("expected index to be NumberExpr, got %T", access.Index)
	}

	if index.Value != "2" {
		t.Fatalf("expected index 2, got %q", index.Value)
	}
}

func TestParseParenthesizedExpression(t *testing.T) {
	tokens := []runner.Token{
		{Type: "OP", Value: "("},
		{Type: "INT", Value: "1"},
		{Type: "OP", Value: "+"},
		{Type: "INT", Value: "2"},
		{Type: "OP", Value: ")"},
		{Type: "OP", Value: "*"},
		{Type: "INT", Value: "3"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	expr := p.parseExpr()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	mul, ok := expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}

	if mul.Op != "*" {
		t.Fatalf("expected top-level '*', got %q", mul.Op)
	}

	add, ok := mul.Left.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected left operand to be BinaryExpr, got %T", mul.Left)
	}

	if add.Op != "+" {
		t.Fatalf("expected parenthesized expression to contain '+', got %q", add.Op)
	}
}

func TestParseBlock(t *testing.T) {
	tokens := []runner.Token{
		{Type: "OP", Value: "{"},

		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "="},
		{Type: "INT", Value: "1"},
		{Type: "OP", Value: ";"},

		{Type: "KEYWORD", Value: "assert"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ";"},

		{Type: "OP", Value: "}"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	statement := p.parseStatement()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	block, ok := statement.(*ast.Block)
	if !ok {
		t.Fatalf("expected Block, got %T", statement)
	}

	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}

	if _, ok := block.Statements[0].(*ast.Assignment); !ok {
		t.Fatalf("expected first statement to be Assignment, got %T",
			block.Statements[0])
	}

	if _, ok := block.Statements[1].(*ast.Assert); !ok {
		t.Fatalf("expected second statement to be Assert, got %T",
			block.Statements[1])
	}
}

func TestParseAssignment(t *testing.T) {
	tokens := []runner.Token{
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "="},
		{Type: "INT", Value: "42"},
		{Type: "OP", Value: ";"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	statement := p.parseAssignment()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	assignment, ok := statement.(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", statement)
	}

	if len(assignment.Variables) != 1 {
		t.Fatalf("expected 1 assignment variable, got %d",
			len(assignment.Variables))
	}

	if len(assignment.Value) != 1 {
		t.Fatalf("expected 1 assignment value, got %d",
			len(assignment.Value))
	}

	variable, ok := assignment.Variables[0].(*ast.VariableExpr)
	if !ok {
		t.Fatalf("expected VariableExpr, got %T", assignment.Variables[0])
	}

	if variable.Name != "x" {
		t.Fatalf("expected variable x, got %q", variable.Name)
	}

	value, ok := assignment.Value[0].(*ast.NumberExpr)
	if !ok {
		t.Fatalf("expected NumberExpr, got %T", assignment.Value[0])
	}

	if value.Value != "42" {
		t.Fatalf("expected value 42, got %q", value.Value)
	}
}

func TestParseArrayAssignment(t *testing.T) {
	tokens := []runner.Token{
		{Type: "IDENT", Value: "a"},
		{Type: "OP", Value: "["},
		{Type: "INT", Value: "1"},
		{Type: "OP", Value: "]"},
		{Type: "OP", Value: "="},
		{Type: "INT", Value: "42"},
		{Type: "OP", Value: ";"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	statement := p.parseAssignment()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	assignment, ok := statement.(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", statement)
	}

	target, ok := assignment.Variables[0].(*ast.ArrayAccessExpr)
	if !ok {
		t.Fatalf("expected ArrayAccessExpr, got %T",
			assignment.Variables[0])
	}

	if target.Array != "a" {
		t.Fatalf("expected array name a, got %q", target.Array)
	}

	index, ok := target.Index.(*ast.NumberExpr)
	if !ok {
		t.Fatalf("expected index NumberExpr, got %T", target.Index)
	}

	if index.Value != "1" {
		t.Fatalf("expected index 1, got %q", index.Value)
	}
}

func TestParseAssert(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "assert"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ";"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	statement := p.parseStatement()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	assertStmt, ok := statement.(*ast.Assert)
	if !ok {
		t.Fatalf("expected Assert, got %T", statement)
	}

	predicate, ok := assertStmt.Predicate.(*ast.BoolPredicate)
	if !ok {
		t.Fatalf("expected BoolPredicate, got %T",
			assertStmt.Predicate)
	}

	if !predicate.Value {
		t.Fatal("expected predicate to be true")
	}
}

func TestParseAssume(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "assume"},
		{Type: "KEYWORD", Value: "false"},
		{Type: "OP", Value: ";"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	statement := p.parseStatement()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	assumeStmt, ok := statement.(*ast.Assume)
	if !ok {
		t.Fatalf("expected Assume, got %T", statement)
	}

	predicate, ok := assumeStmt.Predicate.(*ast.BoolPredicate)
	if !ok {
		t.Fatalf("expected BoolPredicate, got %T",
			assumeStmt.Predicate)
	}

	if predicate.Value {
		t.Fatal("expected predicate to be false")
	}
}

func TestParseIf(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "if"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "=="},
		{Type: "INT", Value: "5"},
		{Type: "OP", Value: ")"},
		{Type: "KEYWORD", Value: "assert"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ";"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	statement := p.parseStatement()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	ifStmt, ok := statement.(*ast.If)
	if !ok {
		t.Fatalf("expected If, got %T", statement)
	}

	if ifStmt.HasElse {
		t.Fatal("expected if without else")
	}

	comparison, ok := ifStmt.Condition.(*ast.Comparison)
	if !ok {
		t.Fatalf("expected Comparison condition, got %T",
			ifStmt.Condition)
	}

	if comparison.Op != "==" {
		t.Fatalf("expected comparison operator ==, got %q",
			comparison.Op)
	}

	assertStmt, ok := ifStmt.Then.(*ast.Assert)
	if !ok {
		t.Fatalf("expected Then to be Assert, got %T",
			ifStmt.Then)
	}

	predicate, ok := assertStmt.Predicate.(*ast.BoolPredicate)
	if !ok {
		t.Fatalf("expected BoolPredicate, got %T",
			assertStmt.Predicate)
	}

	if !predicate.Value {
		t.Fatal("expected assert true")
	}
}

func TestParseIfElse(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "if"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "=="},
		{Type: "INT", Value: "5"},
		{Type: "OP", Value: ")"},

		{Type: "KEYWORD", Value: "assert"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ";"},

		{Type: "KEYWORD", Value: "else"},

		{Type: "KEYWORD", Value: "assume"},
		{Type: "KEYWORD", Value: "false"},
		{Type: "OP", Value: ";"},

		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	statement := p.parseStatement()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	ifStmt, ok := statement.(*ast.If)
	if !ok {
		t.Fatalf("expected If, got %T", statement)
	}

	if !ifStmt.HasElse {
		t.Fatal("expected if to have else")
	}

	if ifStmt.Else == nil {
		t.Fatal("expected Else statement to be non-nil")
	}

	if _, ok := ifStmt.Then.(*ast.Assert); !ok {
		t.Fatalf("expected Then to be Assert, got %T",
			ifStmt.Then)
	}

	if _, ok := ifStmt.Else.(*ast.Assume); !ok {
		t.Fatalf("expected Else to be Assume, got %T",
			ifStmt.Else)
	}
}

func TestParseWhile(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "while"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "<"},
		{Type: "INT", Value: "10"},
		{Type: "OP", Value: ")"},

		{Type: "KEYWORD", Value: "assert"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ";"},

		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	statement := p.parseStatement()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	whileStmt, ok := statement.(*ast.While)
	if !ok {
		t.Fatalf("expected While, got %T", statement)
	}

	if whileStmt.Invariant != nil {
		t.Fatal("expected no invariant")
	}

	comparison, ok := whileStmt.Condition.(*ast.Comparison)
	if !ok {
		t.Fatalf("expected Comparison condition, got %T",
			whileStmt.Condition)
	}

	if comparison.Op != "<" {
		t.Fatalf("expected comparison operator <, got %q",
			comparison.Op)
	}

	if _, ok := whileStmt.Body.(*ast.Assert); !ok {
		t.Fatalf("expected body to be Assert, got %T",
			whileStmt.Body)
	}
}

func TestParseWhileWithInvariant(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "while"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: "<"},
		{Type: "INT", Value: "10"},
		{Type: "OP", Value: ")"},

		{Type: "KEYWORD", Value: "invariant"},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ">="},
		{Type: "INT", Value: "0"},

		{Type: "KEYWORD", Value: "assert"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ";"},

		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	statement := p.parseStatement()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	whileStmt, ok := statement.(*ast.While)
	if !ok {
		t.Fatalf("expected While, got %T", statement)
	}

	if whileStmt.Invariant == nil {
		t.Fatal("expected invariant")
	}

	invariant, ok := (*whileStmt.Invariant).(*ast.PredicateComparison)
	if !ok {
		t.Fatalf("expected PredicateComparison, got %T",
			*whileStmt.Invariant)
	}

	if invariant.Op != ">=" {
		t.Fatalf("expected invariant operator >=, got %q",
			invariant.Op)
	}

	if _, ok := whileStmt.Body.(*ast.Assert); !ok {
		t.Fatalf("expected body to be Assert, got %T",
			whileStmt.Body)
	}
}

func TestParseForallQuantifier(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "forall"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "KEYWORD", Value: "int"},
		{Type: "OP", Value: "|"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ")"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	predicate := p.parsePredicate()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	quantifier, ok := predicate.(*ast.Quantifier)
	if !ok {
		t.Fatalf("expected Quantifier, got %T", predicate)
	}

	if quantifier.Kind != "forall" {
		t.Fatalf("expected forall, got %q", quantifier.Kind)
	}

	if quantifier.Variable.Name != "x" {
		t.Fatalf("expected variable x, got %q",
			quantifier.Variable.Name)
	}

	if quantifier.Variable.Type != ast.IntType {
		t.Fatalf("expected int type, got %q",
			quantifier.Variable.Type)
	}

	body, ok := quantifier.Predicate.(*ast.BoolPredicate)
	if !ok {
		t.Fatalf("expected BoolPredicate, got %T",
			quantifier.Predicate)
	}

	if !body.Value {
		t.Fatal("expected predicate body to be true")
	}
}

func TestParseExistsQuantifier(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "exists"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "a"},
		{Type: "OP", Value: ":"},
		{Type: "KEYWORD", Value: "int[]"},
		{Type: "OP", Value: "|"},
		{Type: "KEYWORD", Value: "false"},
		{Type: "OP", Value: ")"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	predicate := p.parsePredicate()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	quantifier, ok := predicate.(*ast.Quantifier)
	if !ok {
		t.Fatalf("expected Quantifier, got %T", predicate)
	}

	if quantifier.Kind != "exists" {
		t.Fatalf("expected exists, got %q", quantifier.Kind)
	}

	if quantifier.Variable.Name != "a" {
		t.Fatalf("expected variable a, got %q",
			quantifier.Variable.Name)
	}

	if quantifier.Variable.Type != ast.IntArray {
		t.Fatalf("expected int[] type, got %q",
			quantifier.Variable.Type)
	}

	body, ok := quantifier.Predicate.(*ast.BoolPredicate)
	if !ok {
		t.Fatalf("expected BoolPredicate, got %T",
			quantifier.Predicate)
	}

	if body.Value {
		t.Fatal("expected predicate body to be false")
	}
}

func TestParseQuantifierWithComparison(t *testing.T) {
	tokens := []runner.Token{
		{Type: "KEYWORD", Value: "forall"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "KEYWORD", Value: "int"},
		{Type: "OP", Value: "|"},

		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ">"},
		{Type: "INT", Value: "0"},

		{Type: "OP", Value: ")"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	predicate := p.parsePredicate()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	quantifier, ok := predicate.(*ast.Quantifier)
	if !ok {
		t.Fatalf("expected Quantifier, got %T", predicate)
	}

	if quantifier.Kind != "forall" {
		t.Fatalf("expected forall, got %q", quantifier.Kind)
	}

	if quantifier.Variable.Name != "x" {
		t.Fatalf("expected variable x, got %q",
			quantifier.Variable.Name)
	}

	if quantifier.Variable.Type != ast.IntType {
		t.Fatalf("expected int type, got %q",
			quantifier.Variable.Type)
	}

	comparison, ok := quantifier.Predicate.(*ast.PredicateComparison)
	if !ok {
		t.Fatalf("expected PredicateComparison, got %T",
			quantifier.Predicate)
	}

	if comparison.Op != ">" {
		t.Fatalf("expected comparison operator >, got %q",
			comparison.Op)
	}

	left, ok := comparison.Left.(*ast.VariableExpr)
	if !ok {
		t.Fatalf("expected left side to be VariableExpr, got %T",
			comparison.Left)
	}

	if left.Name != "x" {
		t.Fatalf("expected left variable x, got %q", left.Name)
	}

	right, ok := comparison.Right.(*ast.NumberExpr)
	if !ok {
		t.Fatalf("expected right side to be NumberExpr, got %T",
			comparison.Right)
	}

	if right.Value != "0" {
		t.Fatalf("expected right value 0, got %q", right.Value)
	}
}

func TestParseFormulaRef(t *testing.T) {
	tokens := []runner.Token{
		{Type: "IDENT", Value: "positive"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ")"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)

	predicate := p.parsePredicate()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.errors)
	}

	ref, ok := predicate.(*ast.FormulaRef)
	if !ok {
		t.Fatalf("expected FormulaRef, got %T", predicate)
	}

	if ref.Name != "positive" {
		t.Fatalf("expected formula name positive, got %q", ref.Name)
	}

	if len(ref.Args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(ref.Args))
	}

	arg, ok := ref.Args[0].(*ast.VariableExpr)
	if !ok {
		t.Fatalf("expected VariableExpr argument, got %T", ref.Args[0])
	}

	if arg.Name != "x" {
		t.Fatalf("expected argument x, got %q", arg.Name)
	}
}

func TestParseFormula(t *testing.T) {
	tokens := []runner.Token{
		{Type: "IDENT", Value: "positive"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "KEYWORD", Value: "int"},
		{Type: "OP", Value: ")"},
		{Type: "OP", Value: "=>"},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ">"},
		{Type: "INT", Value: "0"},
		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	formula := p.parseFormula()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parse errors: %+v", p.errors)
	}

	if formula.Name != "positive" {
		t.Fatalf("expected formula name positive, got %q", formula.Name)
	}

	if len(formula.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(formula.Parameters))
	}

	if formula.Parameters[0].Name != "x" {
		t.Fatalf("expected parameter name x, got %q", formula.Parameters[0].Name)
	}

	if formula.Parameters[0].Type != ast.IntType {
		t.Fatalf("expected parameter type int, got %q", formula.Parameters[0].Type)
	}

	if formula.Body == nil {
		t.Fatal("expected formula body")
	}

	comparison, ok := formula.Body.(*ast.PredicateComparison)
	if !ok {
		t.Fatalf("expected PredicateComparison, got %T", formula.Body)
	}

	if comparison.Op != ">" {
		t.Fatalf("expected operator >, got %q", comparison.Op)
	}
}

func TestParseProgram(t *testing.T) {
	tokens := []runner.Token{
		// formula positive(x: int) => x > 0
		{Type: "IDENT", Value: "positive"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "KEYWORD", Value: "int"},
		{Type: "OP", Value: ")"},
		{Type: "OP", Value: "=>"},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ">"},
		{Type: "INT", Value: "0"},

		// function foo(x: int) returns y: int { assert true; }
		{Type: "IDENT", Value: "foo"},
		{Type: "OP", Value: "("},
		{Type: "IDENT", Value: "x"},
		{Type: "OP", Value: ":"},
		{Type: "KEYWORD", Value: "int"},
		{Type: "OP", Value: ")"},
		{Type: "KEYWORD", Value: "returns"},
		{Type: "IDENT", Value: "y"},
		{Type: "OP", Value: ":"},
		{Type: "KEYWORD", Value: "int"},
		{Type: "OP", Value: "{"},
		{Type: "KEYWORD", Value: "assert"},
		{Type: "KEYWORD", Value: "true"},
		{Type: "OP", Value: ";"},
		{Type: "OP", Value: "}"},

		{Type: "EOF", Value: ""},
	}

	p := New(tokens)
	program := p.parseProgram()

	if len(p.errors) != 0 {
		t.Fatalf("unexpected parse errors: %+v", p.errors)
	}

	if len(program.Formulas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(program.Formulas))
	}

	if program.Formulas[0].Name != "positive" {
		t.Fatalf("expected formula positive, got %q",
			program.Formulas[0].Name)
	}

	if len(program.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(program.Functions))
	}

	if program.Functions[0].Name != "foo" {
		t.Fatalf("expected function foo, got %q",
			program.Functions[0].Name)
	}
}
