package semantic

import (
	"fmt"
	"funny-lexer/internal/ast"
	"strconv"
)

func (a *Analyzer) checkExpr(x ast.Expr) Type {
	if x == nil {
		return Invalid
	}
	var t Type
	switch e := x.(type) {
	case *ast.NumberExpr:
		if _, err := strconv.ParseInt(e.Value, 10, 64); err != nil {
			a.add(fmt.Sprintf("invalid integer literal %q", e.Value))
			t = Invalid
		} else {
			t = Int
		}
	case ast.NumberExpr:
		return a.checkExpr(&e)
	case *ast.VariableExpr:
		s, ok := a.scope.lookup(e.Name)
		if !ok {
			a.add(fmt.Sprintf("undefined variable %q", e.Name))
			t = Invalid
		} else if s.Kind != VariableSymbol {
			a.add(fmt.Sprintf("%q is not a variable", e.Name))
			t = Invalid
		} else {
			t = s.Type
		}
	case ast.VariableExpr:
		return a.checkExpr(&e)
	case *ast.UnaryExpr:
		v := a.checkExpr(e.X)
		if e.Op != "-" {
			a.add(fmt.Sprintf("unsupported unary operator %q", e.Op))
			t = Invalid
		} else if !compatible(Int, v) {
			a.add(fmt.Sprintf("unary - requires int, got %s", v))
			t = Invalid
		} else {
			t = Int
		}
	case ast.UnaryExpr:
		return a.checkExpr(&e)
	case *ast.BinaryExpr:
		l, r := a.checkExpr(e.Left), a.checkExpr(e.Right)
		if e.Op == "+" || e.Op == "-" || e.Op == "*" || e.Op == "/" {
			if !compatible(Int, l) || !compatible(Int, r) {
				a.add(fmt.Sprintf("operator %s requires int operands, got %s and %s", e.Op, l, r))
				t = Invalid
			} else {
				t = Int
			}
		} else {
			a.add(fmt.Sprintf("unsupported binary operator %q", e.Op))
			t = Invalid
		}
	case ast.BinaryExpr:
		return a.checkExpr(&e)
	case *ast.ArrayAccessExpr:
		s, ok := a.scope.lookup(e.Array)
		if !ok {
			a.add(fmt.Sprintf("undefined array %q", e.Array))
			t = Invalid
		} else if s.Type != IntArray && s.Type != Unknown {
			a.add(fmt.Sprintf("%q is %s, not an array", e.Array, s.Type))
			t = Invalid
		} else {
			t = Int
		}
		idx := a.checkExpr(e.Index)
		if !compatible(Int, idx) {
			a.add(fmt.Sprintf("array index must be int, got %s", idx))
			t = Invalid
		}
	case ast.ArrayAccessExpr:
		return a.checkExpr(&e)
	case *ast.CallExpr:
		t = a.checkCall(e.Name, e.Args)
	case ast.CallExpr:
		return a.checkExpr(&e)
	default:
		a.add(fmt.Sprintf("unsupported expression node %T", x))
		t = Invalid
	}
	a.result.Types[x] = t
	return t
}
func (a *Analyzer) checkCall(name string, args []ast.Expr) Type {
	s, ok := a.global.lookup(name)
	if !ok {
		a.add(fmt.Sprintf("undefined function or formula %q", name))
		for _, x := range args {
			a.checkExpr(x)
		}
		return Invalid
	}
	if s.Kind != FunctionSymbol && s.Kind != FormulaSymbol {
		a.add(fmt.Sprintf("%q is not callable", name))
		return Invalid
	}
	if len(args) != len(s.Parameters) {
		a.add(fmt.Sprintf("call %q expects %d arguments, got %d", name, len(s.Parameters), len(args)))
	}
	for i, x := range args {
		got := a.checkExpr(x)
		if i < len(s.Parameters) && !compatible(s.Parameters[i], got) {
			a.add(fmt.Sprintf("argument %d of %q: expected %s, got %s", i+1, name, s.Parameters[i], got))
		}
	}
	if s.Kind == FormulaSymbol {
		return Bool
	}
	if len(s.Returns) == 0 {
		return Unknown
	}
	if len(s.Returns) > 1 {
		a.add(fmt.Sprintf("function %q returns multiple values; use multiple assignment", name))
		return Invalid
	}
	return s.Returns[0]
}
