package semantic

import (
	"strings"
	"testing"

	"funny-lexer/internal/ast"
)

func intVar(name string) ast.Variable { return ast.Variable{Name: name, Type: ast.IntType} }
func number(value string) ast.Expr    { return &ast.NumberExpr{Value: value} }
func variable(name string) ast.Expr   { return &ast.VariableExpr{Name: name} }

func hasError(result *Result, fragment string) bool {
	for _, err := range result.Errors {
		if strings.Contains(err.Message, fragment) {
			return true
		}
	}
	return false
}

func TestAnalyzeValidFunction(t *testing.T) {
	program := &ast.Program{Functions: []ast.Function{{
		Name: "increment", Parameters: []ast.Variable{intVar("x")}, Returns: []ast.Variable{intVar("result")},
		Body: &ast.Block{Statements: []ast.Statement{&ast.Assignment{
			Variables: []ast.Expr{variable("result")},
			Value:     []ast.Expr{&ast.BinaryExpr{Left: variable("x"), Op: "+", Right: number("1")}},
		}}},
	}}}
	if result := Analyze(program); len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got: %+v", result.Errors)
	}
}

func TestAnalyzeUndefinedVariable(t *testing.T) {
	program := &ast.Program{Functions: []ast.Function{{Name: "f", Returns: []ast.Variable{intVar("result")},
		Body: &ast.Block{Statements: []ast.Statement{&ast.Assignment{Variables: []ast.Expr{variable("result")}, Value: []ast.Expr{variable("missing")}}}},
	}}}
	result := Analyze(program)
	if !hasError(result, `undefined variable "missing"`) {
		t.Fatalf("expected undefined-variable error, got: %+v", result.Errors)
	}
}

func TestAnalyzeDuplicateParameter(t *testing.T) {
	program := &ast.Program{Functions: []ast.Function{{Name: "f", Parameters: []ast.Variable{intVar("x"), intVar("x")}, Body: &ast.Block{}}}}
	result := Analyze(program)
	if !hasError(result, `duplicate declaration of "x"`) {
		t.Fatalf("expected duplicate declaration error, got: %+v", result.Errors)
	}
}

func TestAnalyzeInvalidIntegerLiteral(t *testing.T) {
	program := &ast.Program{Functions: []ast.Function{{Name: "f", Returns: []ast.Variable{intVar("result")},
		Body: &ast.Block{Statements: []ast.Statement{&ast.Assignment{Variables: []ast.Expr{variable("result")}, Value: []ast.Expr{number("not-an-integer")}}}},
	}}}
	result := Analyze(program)
	if !hasError(result, "invalid integer literal") {
		t.Fatalf("expected invalid integer literal error, got: %+v", result.Errors)
	}
}

func TestAnalyzeFunctionCallArgumentCount(t *testing.T) {
	program := &ast.Program{Functions: []ast.Function{
		{Name: "callee", Parameters: []ast.Variable{intVar("x")}, Returns: []ast.Variable{intVar("result")}, Body: &ast.Block{}},
		{Name: "caller", Returns: []ast.Variable{intVar("result")}, Body: &ast.Block{Statements: []ast.Statement{&ast.Assignment{
			Variables: []ast.Expr{variable("result")}, Value: []ast.Expr{&ast.CallExpr{Name: "callee"}},
		}}}},
	}}
	result := Analyze(program)
	if !hasError(result, `call "callee" expects 1 arguments, got 0`) {
		t.Fatalf("expected argument-count error, got: %+v", result.Errors)
	}
}

func TestAnalyzeDuplicateTopLevelDeclaration(t *testing.T) {
	program := &ast.Program{Functions: []ast.Function{{Name: "same", Body: &ast.Block{}}, {Name: "same", Body: &ast.Block{}}}}
	result := Analyze(program)
	if !hasError(result, `duplicate declaration of "same"`) {
		t.Fatalf("expected duplicate top-level declaration error, got: %+v", result.Errors)
	}
}

func TestAnalyzeFormulaAndQuantifier(t *testing.T) {
	program := &ast.Program{
		Formulas: []ast.Formula{{Name: "positive", Parameters: []ast.Variable{intVar("x")}, Body: ast.PredicateComparison{Left: variable("x"), Op: ">", Right: number("0")}}},
		Functions: []ast.Function{{Name: "f", Body: &ast.Block{Statements: []ast.Statement{&ast.Assert{Predicate: ast.Quantifier{
			Kind: "forall", Variable: intVar("y"), Predicate: ast.FormulaRef{Name: "positive", Args: []ast.Expr{variable("y")}},
		}}}}}},
	}
	result := Analyze(program)
	if len(result.Errors) != 0 {
		t.Fatalf("expected valid formula/quantifier, got: %+v", result.Errors)
	}
}

func TestAnalyzeNilProgram(t *testing.T) {
	result := Analyze(nil)
	if !hasError(result, "nil program") {
		t.Fatalf("expected nil-program error, got: %+v", result.Errors)
	}
}

func TestInferLocalVariableType(t *testing.T) {
	x := variable("x")
	expression := &ast.BinaryExpr{
		Left:  variable("a"),
		Op:    "+",
		Right: number("1"),
	}

	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Body: &ast.Block{
				Statements: []ast.Statement{
					&ast.Assignment{
						Variables: []ast.Expr{x},
						Value:     []ast.Expr{expression},
					},
				},
			},
		}},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got: %+v", result.Errors)
	}

	if got := result.Types[x]; got != Int {
		t.Errorf("expected x to have type int, got %s", got)
	}

	if got := result.Types[expression]; got != Int {
		t.Errorf("expected assignment expression to have type int, got %s", got)
	}
}

func TestInferLocalVariableRejectsConflictingType(t *testing.T) {
	x1 := variable("x")
	x2 := variable("x")

	program := &ast.Program{
		Formulas: []ast.Formula{{
			Name:       "positive",
			Parameters: []ast.Variable{intVar("a")},
			Body: ast.PredicateComparison{
				Left:  variable("a"),
				Op:    ">",
				Right: number("0"),
			},
		}},
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Body: &ast.Block{
				Statements: []ast.Statement{
					&ast.Assignment{
						Variables: []ast.Expr{x1},
						Value: []ast.Expr{
							&ast.BinaryExpr{
								Left:  variable("a"),
								Op:    "+",
								Right: number("1"),
							},
						},
					},
					&ast.Assignment{
						Variables: []ast.Expr{x2},
						Value: []ast.Expr{
							&ast.CallExpr{
								Name: "positive",
								Args: []ast.Expr{variable("a")},
							},
						},
					},
				},
			},
		}},
	}

	result := Analyze(program)

	if !hasError(result, `cannot assign bool to "x" of type int`) {
		t.Fatalf("expected conflicting-type error, got: %+v", result.Errors)
	}

	if got := result.Types[x1]; got != Int {
		t.Errorf("expected first assignment target to have type int, got %s", got)
	}

	if got := result.Types[x2]; got != Int {
		t.Errorf("expected conflicting assignment target to retain type int, got %s", got)
	}
}

func TestInferLocalVariableAfterInvalidExpression(t *testing.T) {
	x1 := variable("x")
	x2 := variable("x")

	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Body: &ast.Block{
				Statements: []ast.Statement{
					&ast.Assignment{
						Variables: []ast.Expr{x1},
						Value:     []ast.Expr{variable("missing")},
					},
					&ast.Assignment{
						Variables: []ast.Expr{x2},
						Value: []ast.Expr{
							&ast.BinaryExpr{
								Left:  variable("a"),
								Op:    "+",
								Right: number("1"),
							},
						},
					},
				},
			},
		}},
	}

	result := Analyze(program)

	if !hasError(result, `undefined variable "missing"`) {
		t.Fatalf("expected undefined-variable error, got: %+v", result.Errors)
	}

	if got := result.Types[x1]; got != Unknown {
		t.Errorf("expected first assignment target to remain unknown, got %s", got)
	}

	if got := result.Types[x2]; got != Int {
		t.Errorf("expected later assignment target to have type int, got %s", got)
	}
}

func TestInferredTypeIsPreservedOnLaterAssignments(t *testing.T) {
	x1 := variable("x")
	x2 := variable("x")

	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Body: &ast.Block{
				Statements: []ast.Statement{
					&ast.Assignment{
						Variables: []ast.Expr{x1},
						Value: []ast.Expr{
							&ast.BinaryExpr{
								Left:  variable("a"),
								Op:    "+",
								Right: number("1"),
							},
						},
					},
					&ast.Assignment{
						Variables: []ast.Expr{x2},
						Value:     []ast.Expr{variable("x")},
					},
				},
			},
		}},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got: %+v", result.Errors)
	}

	if got := result.Types[x1]; got != Int {
		t.Errorf("expected first target to have type int, got %s", got)
	}

	if got := result.Types[x2]; got != Int {
		t.Errorf("expected second target to have type int, got %s", got)
	}
}
