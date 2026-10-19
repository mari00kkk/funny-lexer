package semantic

import (
	"fmt"
	"funny-lexer/internal/ast"
)

func (a *Analyzer) checkStatement(st ast.Statement) {
	if st == nil {
		return
	}
	switch s := st.(type) {
	case *ast.Block:
		old := a.scope
		a.scope = newScope(old)
		for _, x := range s.Statements {
			a.checkStatement(x)
		}
		a.scope = old
	case ast.Block:
		a.checkStatement(&s)
	case *ast.Assignment:
		a.checkAssignment(s)
	case ast.Assignment:
		a.checkAssignment(&s)
	case *ast.If:
		a.checkCondition(s.Condition)
		a.checkStatement(s.Then)
		if s.HasElse {
			a.checkStatement(s.Else)
		}
	case ast.If:
		a.checkStatement(&s)
	case *ast.While:
		a.checkCondition(s.Condition)
		if s.Invariant != nil {
			a.checkPredicate(*s.Invariant)
		}
		a.checkStatement(s.Body)
	case ast.While:
		a.checkStatement(&s)
	case *ast.Assert:
		a.checkPredicate(s.Predicate)
	case ast.Assert:
		a.checkPredicate(s.Predicate)
	case *ast.Assume:
		a.checkPredicate(s.Predicate)
	case ast.Assume:
		a.checkPredicate(s.Predicate)
	default:
		a.add(fmt.Sprintf("unsupported statement node %T", st))
	}
}
func (a *Analyzer) checkAssignment(s *ast.Assignment) {
	vals := make([]Type, len(s.Value))
	multiReturns := []Type(nil)
	if len(s.Variables) > 1 && len(s.Value) == 1 {
		if call, ok := s.Value[0].(*ast.CallExpr); ok {
			sym, found := a.global.lookup(call.Name)
			if found && sym.Kind == FunctionSymbol { // Check call arguments without treating its multi-return as a scalar.
				if len(call.Args) != len(sym.Parameters) {
					a.add(fmt.Sprintf("call %q expects %d arguments, got %d", call.Name, len(sym.Parameters), len(call.Args)))
				}
				for i, arg := range call.Args {
					got := a.checkExpr(arg)
					if i < len(sym.Parameters) && !compatible(sym.Parameters[i], got) {
						a.add(fmt.Sprintf("argument %d of %q: expected %s, got %s", i+1, call.Name, sym.Parameters[i], got))
					}
				}
				multiReturns = sym.Returns
			} else {
				vals[0] = a.checkExpr(s.Value[0])
			}
		} else {
			vals[0] = a.checkExpr(s.Value[0])
		}
	} else {
		for i, v := range s.Value {
			vals[i] = a.checkExpr(v)
		}
	}
	if multiReturns != nil {
		if len(multiReturns) != len(s.Variables) {
			a.add(fmt.Sprintf("function returns %d values, assignment has %d targets", len(multiReturns), len(s.Variables)))
		}
		for i, x := range s.Variables {
			if i < len(multiReturns) {
				a.assignTarget(x, multiReturns[i])
			}
		}
		return
	}
	if len(s.Variables) != len(vals) {
		a.add(fmt.Sprintf("assignment has %d targets and %d values", len(s.Variables), len(vals)))
		return
	}
	for i, x := range s.Variables {
		a.assignTarget(x, vals[i])
	}
}
func (a *Analyzer) assignTarget(target ast.Expr, value Type) {
	switch t := target.(type) {
	case *ast.VariableExpr:
		a.assignVariable(t.Name, value)

		if s, ok := a.scope.lookup(t.Name); ok {
			a.result.Types[target] = s.Type
		}

	case ast.VariableExpr:
		a.assignVariable(t.Name, value)

		if s, ok := a.scope.lookup(t.Name); ok {
			a.result.Types[target] = s.Type
		}

	case *ast.ArrayAccessExpr:
		s, ok := a.scope.lookup(t.Array)
		if !ok {
			a.add(fmt.Sprintf("undefined array %q", t.Array))
		} else if s.Type != IntArray && s.Type != Unknown {
			a.add(fmt.Sprintf(
				"cannot index non-array %q of type %s",
				t.Array, s.Type,
			))
		}

		idx := a.checkExpr(t.Index)
		if !compatible(Int, idx) {
			a.add(fmt.Sprintf("array index must be int, got %s", idx))
		}

		if !compatible(Int, value) {
			a.add(fmt.Sprintf(
				"array element assignment expects int, got %s",
				value,
			))
		}

	case ast.ArrayAccessExpr:
		a.assignTarget(&t, value)

	default:
		a.add(fmt.Sprintf("invalid assignment target %T", target))
	}
}
func (a *Analyzer) assignVariable(name string, value Type) {
	if name == "" {
		return
	}

	s, ok := a.scope.lookup(name)

	if !ok {
		// Нельзя вывести тип из ошибочного или неизвестного выражения.
		inferredType := value
		if value == Invalid || value == Unknown {
			inferredType = Unknown
		}

		a.scope.Symbols[name] = Symbol{
			Name:     name,
			Kind:     VariableSymbol,
			Type:     inferredType,
			Inferred: true,
		}
		return
	}

	if s.Kind != VariableSymbol {
		a.add(fmt.Sprintf("%q is not assignable", name))
		return
	}

	if s.Type == Unknown {
		// Выводим тип только из корректного выражения,
		// тип которого действительно известен.
		if value == Unknown || value == Invalid {
			return
		}

		s.Type = value
		s.Inferred = true

		for cur := a.scope; cur != nil; cur = cur.Parent {
			if old, exists := cur.Symbols[name]; exists &&
				old.Kind == VariableSymbol {
				cur.Symbols[name] = s
				break
			}
		}

		return
	}

	if !compatible(s.Type, value) {
		a.add(fmt.Sprintf(
			"cannot assign %s to %q of type %s",
			value, name, s.Type,
		))
	}
}
func (a *Analyzer) checkCondition(c ast.Condition) Type {
	if c == nil {
		return Invalid
	}
	switch x := c.(type) {
	case *ast.BoolCondition:
		return Bool
	case ast.BoolCondition:
		return Bool
	case *ast.Comparison:
		l, r := a.checkExpr(x.Left), a.checkExpr(x.Right)
		if !compatible(l, r) {
			a.add(fmt.Sprintf("comparison %s has incompatible operands %s and %s", x.Op, l, r))
		}
		if (x.Op == "<" || x.Op == "<=" || x.Op == ">" || x.Op == ">=") && (l != Int || r != Int) {
			a.add(fmt.Sprintf("operator %s requires int operands", x.Op))
		}
		return Bool
	case ast.Comparison:
		return a.checkCondition(&x)
	case *ast.NotCondition:
		if t := a.checkCondition(x.X); !compatible(Bool, t) {
			a.add(fmt.Sprintf("not requires bool, got %s", t))
		}
		return Bool
	case ast.NotCondition:
		return a.checkCondition(&x)
	case *ast.BinaryCondition:
		l, r := a.checkCondition(x.Left), a.checkCondition(x.Right)
		if !compatible(Bool, l) || !compatible(Bool, r) {
			a.add(fmt.Sprintf("logical operator %s requires bool operands", x.Op))
		}
		return Bool
	case ast.BinaryCondition:
		return a.checkCondition(&x)
	default:
		a.add(fmt.Sprintf("unsupported condition node %T", c))
		return Invalid
	}
}
func (a *Analyzer) checkPredicate(p ast.Predicate) Type {
	if p == nil {
		return Invalid
	}
	switch x := p.(type) {
	case *ast.BoolPredicate:
		return Bool
	case ast.BoolPredicate:
		return Bool
	case *ast.PredicateComparison:
		l, r := a.checkExpr(x.Left), a.checkExpr(x.Right)
		if !compatible(l, r) {
			a.add(fmt.Sprintf("predicate comparison %s has incompatible operands %s and %s", x.Op, l, r))
		}
		if (x.Op == "<" || x.Op == "<=" || x.Op == ">" || x.Op == ">=") && (l != Int || r != Int) {
			a.add(fmt.Sprintf("operator %s requires int operands", x.Op))
		}
		return Bool
	case ast.PredicateComparison:
		return a.checkPredicate(&x)
	case *ast.NotPredicate:
		if t := a.checkPredicate(x.X); !compatible(Bool, t) {
			a.add(fmt.Sprintf("not requires bool, got %s", t))
		}
		return Bool
	case ast.NotPredicate:
		return a.checkPredicate(&x)
	case *ast.BinaryPredicate:
		l, r := a.checkPredicate(x.Left), a.checkPredicate(x.Right)
		if !compatible(Bool, l) || !compatible(Bool, r) {
			a.add(fmt.Sprintf("logical operator %s requires bool operands", x.Op))
		}
		return Bool
	case ast.BinaryPredicate:
		return a.checkPredicate(&x)
	case *ast.Quantifier:
		old := a.scope
		a.scope = newScope(old)
		a.declare(x.Variable, "quantifier variable")
		t := a.checkPredicate(x.Predicate)
		if !compatible(Bool, t) {
			a.add("quantifier body must be bool")
		}
		a.scope = old
		return Bool
	case ast.Quantifier:
		return a.checkPredicate(&x)
	case *ast.FormulaRef:
		s, ok := a.global.lookup(x.Name)
		if !ok {
			a.add(fmt.Sprintf("undefined formula %q", x.Name))
			for _, arg := range x.Args {
				a.checkExpr(arg)
			}
			return Invalid
		}
		if s.Kind != FormulaSymbol {
			a.add(fmt.Sprintf("%q is not a formula", x.Name))
			return Invalid
		}
		if len(x.Args) != len(s.Parameters) {
			a.add(fmt.Sprintf("formula %q expects %d arguments, got %d", x.Name, len(s.Parameters), len(x.Args)))
		}
		for i, arg := range x.Args {
			got := a.checkExpr(arg)
			if i < len(s.Parameters) && !compatible(s.Parameters[i], got) {
				a.add(fmt.Sprintf("argument %d of formula %q: expected %s, got %s", i+1, x.Name, s.Parameters[i], got))
			}
		}
		return Bool
	case ast.FormulaRef:
		return a.checkPredicate(&x)
	default:
		a.add(fmt.Sprintf("unsupported predicate node %T", p))
		return Invalid
	}
}
