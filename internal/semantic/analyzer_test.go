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

func TestAnalyzeUnaryExpression(t *testing.T) {
	expression := &ast.UnaryExpr{
		Op: "-",
		X:  variable("a"),
	}

	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Returns:    []ast.Variable{intVar("result")},
			Body: &ast.Block{Statements: []ast.Statement{
				&ast.Assignment{
					Variables: []ast.Expr{variable("result")},
					Value:     []ast.Expr{expression},
				},
			}},
		}},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors for unary minus, got: %+v", result.Errors)
	}

	if got := result.Types[expression]; got != Int {
		t.Errorf("expected unary expression type int, got %s", got)
	}
}

func TestAnalyzeArrayAccess(t *testing.T) {
	expression := &ast.ArrayAccessExpr{
		Array: "arr",
		Index: variable("i"),
	}

	program := &ast.Program{
		Functions: []ast.Function{{
			Name: "f",
			Parameters: []ast.Variable{
				{Name: "arr", Type: ast.IntArray},
				intVar("i"),
			},
			Returns: []ast.Variable{intVar("result")},
			Body: &ast.Block{Statements: []ast.Statement{
				&ast.Assignment{
					Variables: []ast.Expr{variable("result")},
					Value:     []ast.Expr{expression},
				},
			}},
		}},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors for valid array access, got: %+v", result.Errors)
	}

	if got := result.Types[expression]; got != Int {
		t.Errorf("expected array element type int, got %s", got)
	}
}

func TestAnalyzeArrayAssignment(t *testing.T) {
	program := &ast.Program{
		Functions: []ast.Function{{
			Name: "f",
			Parameters: []ast.Variable{
				{Name: "arr", Type: ast.IntArray},
				intVar("i"),
			},
			Body: &ast.Block{Statements: []ast.Statement{
				&ast.Assignment{
					Variables: []ast.Expr{
						&ast.ArrayAccessExpr{
							Array: "arr",
							Index: variable("i"),
						},
					},
					Value: []ast.Expr{number("10")},
				},
			}},
		}},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected valid array element assignment, got: %+v", result.Errors)
	}
}

func TestAnalyzeFunctionArgumentTypeMismatch(t *testing.T) {
	program := &ast.Program{
		Functions: []ast.Function{
			{
				Name:       "callee",
				Parameters: []ast.Variable{intVar("x")},
				Returns:    []ast.Variable{intVar("result")},
				Body:       &ast.Block{},
			},
			{
				Name: "caller",
				Parameters: []ast.Variable{
					{Name: "arr", Type: ast.IntArray},
				},
				Returns: []ast.Variable{intVar("result")},
				Body: &ast.Block{Statements: []ast.Statement{
					&ast.Assignment{
						Variables: []ast.Expr{variable("result")},
						Value: []ast.Expr{
							&ast.CallExpr{
								Name: "callee",
								Args: []ast.Expr{variable("arr")},
							},
						},
					},
				}},
			},
		},
	}

	result := Analyze(program)

	if !hasError(result, `argument 1 of "callee": expected int, got int[]`) {
		t.Fatalf("expected argument type mismatch, got: %+v", result.Errors)
	}
}

func TestAnalyzeAssignmentTargetValueCountMismatch(t *testing.T) {
	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Body: &ast.Block{Statements: []ast.Statement{
				&ast.Assignment{
					Variables: []ast.Expr{
						variable("x"),
						variable("y"),
					},
					Value: []ast.Expr{variable("a")},
				},
			}},
		}},
	}

	result := Analyze(program)

	if !hasError(result, "assignment has 2 targets and 1 values") {
		t.Fatalf("expected assignment count error, got: %+v", result.Errors)
	}
}

func TestAnalyzeMultipleReturnValues(t *testing.T) {
	program := &ast.Program{
		Functions: []ast.Function{
			{
				Name: "pair",
				Returns: []ast.Variable{
					intVar("first"),
					intVar("second"),
				},
				Body: &ast.Block{Statements: []ast.Statement{
					&ast.Assignment{
						Variables: []ast.Expr{
							variable("first"),
							variable("second"),
						},
						Value: []ast.Expr{
							number("1"),
							number("2"),
						},
					},
				}},
			},
			{
				Name: "caller",
				Body: &ast.Block{Statements: []ast.Statement{
					&ast.Assignment{
						Variables: []ast.Expr{
							variable("x"),
							variable("y"),
						},
						Value: []ast.Expr{
							&ast.CallExpr{Name: "pair"},
						},
					},
				}},
			},
		},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected valid multiple-return assignment, got: %+v", result.Errors)
	}
}

func TestAnalyzeIfAndWhileStatements(t *testing.T) {
	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Returns:    []ast.Variable{intVar("result")},
			Body: &ast.Block{Statements: []ast.Statement{
				&ast.If{
					Condition: &ast.Comparison{
						Left:  variable("a"),
						Op:    ">",
						Right: number("0"),
					},
					Then: &ast.Assignment{
						Variables: []ast.Expr{variable("result")},
						Value:     []ast.Expr{variable("a")},
					},
					Else: &ast.Assignment{
						Variables: []ast.Expr{variable("result")},
						Value:     []ast.Expr{number("0")},
					},
					HasElse: true,
				},
				&ast.While{
					Condition: &ast.Comparison{
						Left:  variable("a"),
						Op:    ">",
						Right: number("0"),
					},
					Body: &ast.Assignment{
						Variables: []ast.Expr{variable("result")},
						Value:     []ast.Expr{variable("a")},
					},
				},
			}},
		}},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected valid if and while statements, got: %+v", result.Errors)
	}
}

func TestAnalyzeRequiresAndEnsures(t *testing.T) {
	requires := ast.Predicate(ast.PredicateComparison{
		Left:  variable("a"),
		Op:    ">=",
		Right: number("0"),
	})

	ensures := ast.Predicate(ast.PredicateComparison{
		Left:  variable("result"),
		Op:    ">=",
		Right: number("0"),
	})

	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Returns:    []ast.Variable{intVar("result")},
			Requires:   &requires,
			Ensures:    &ensures,
			Body: &ast.Block{Statements: []ast.Statement{
				&ast.Assignment{
					Variables: []ast.Expr{variable("result")},
					Value:     []ast.Expr{variable("a")},
				},
			}},
		}},
	}

	result := Analyze(program)

	if len(result.Errors) != 0 {
		t.Fatalf("expected valid requires and ensures predicates, got: %+v", result.Errors)
	}
}

func TestMarshalAnnotatedJSON(t *testing.T) {
	expression := &ast.BinaryExpr{
		Left:  variable("a"),
		Op:    "+",
		Right: number("1"),
	}

	program := &ast.Program{
		Functions: []ast.Function{{
			Name:       "f",
			Parameters: []ast.Variable{intVar("a")},
			Returns:    []ast.Variable{intVar("result")},
			Body: &ast.Block{Statements: []ast.Statement{
				&ast.Assignment{
					Variables: []ast.Expr{variable("result")},
					Value:     []ast.Expr{expression},
				},
			}},
		}},
	}

	result := Analyze(program)

	data, err := MarshalAnnotatedJSON(program, result)
	if err != nil {
		t.Fatalf("unexpected JSON serialization error: %v", err)
	}

	output := string(data)

	if !strings.Contains(output, `"ast"`) {
		t.Error("expected JSON to contain the ast field")
	}

	if !strings.Contains(output, `"errors"`) {
		t.Error("expected JSON to contain the errors field")
	}

	if !strings.Contains(output, `"inferredType": "int"`) {
		t.Error("expected JSON to contain an int inferredType annotation")
	}
}

func TestMarshalAnnotatedJSONRejectsNilProgram(t *testing.T) {
	_, err := MarshalAnnotatedJSON(nil, &Result{})

	if err == nil {
		t.Fatal("expected an error when serializing a nil program")
	}
}

func TestMarshalAnnotatedJSONRejectsNilResult(t *testing.T) {
	_, err := MarshalAnnotatedJSON(&ast.Program{}, nil)

	if err == nil {
		t.Fatal("expected an error when serializing a nil semantic result")
	}
}
