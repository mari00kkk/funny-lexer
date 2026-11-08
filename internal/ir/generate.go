package ir

import (
	"fmt"
	"funny-lexer/internal/ast"
	"strconv"
	"strings"
)

type generator struct {
	fn          Function
	temp, label int
}

func (g *generator) tmp() string  { g.temp++; return fmt.Sprintf("%%t%d", g.temp) }
func (g *generator) lab() string  { g.label++; return fmt.Sprintf("L%d", g.label) }
func (g *generator) emit(i Instr) { g.fn.Instructions = append(g.fn.Instructions, i) }
func Generate(p *ast.Program) (*Program, error) {
	if p == nil {
		return nil, fmt.Errorf("nil AST")
	}
	out := &Program{}
	for _, f := range p.Functions {
		g := &generator{}
		g.fn.Name = f.Name
		for _, v := range f.Parameters {
			g.fn.Parameters = append(g.fn.Parameters, v.Name)
			g.fn.ParameterTypes = append(g.fn.ParameterTypes, string(v.Type))
		}
		for _, v := range f.Returns {
			g.fn.Returns = append(g.fn.Returns, v.Name)
			g.fn.ReturnTypes = append(g.fn.ReturnTypes, string(v.Type))
		}
		g.fn.LocalTypes = make(map[string]string)
		for _, v := range f.Uses {
			g.fn.LocalTypes[v.Name] = string(v.Type)
			g.fn.Locals = append(g.fn.Locals, v.Name)
		}

		if err := g.stmt(f.Body); err != nil {
			return nil, fmt.Errorf("function %s: %w", f.Name, err)
		}
		args := append([]string(nil), g.fn.Returns...)
		g.emit(Instr{Op: OpReturn, Args: args})
		g.collectLocals()
		out.Functions = append(out.Functions, g.fn)
	}
	return out, nil
}
func (g *generator) collectLocals() {
	seen := map[string]bool{}
	for _, n := range g.fn.Parameters {
		seen[n] = true
	}
	for _, n := range g.fn.Returns {
		seen[n] = true
	}
	for _, n := range g.fn.Locals {
		seen[n] = true
	}
	add := func(n string) {
		if n == "" || strings.HasPrefix(n, "%") {
			return
		}
		if _, err := strconv.ParseInt(n, 10, 64); err == nil {
			return
		}
		if !seen[n] {
			seen[n] = true
			g.fn.Locals = append(g.fn.Locals, n)
		}
	}
	for _, in := range g.fn.Instructions {
		switch in.Op {
		case OpMove:
			add(in.Dest)
			add(in.Arg1)
		case OpBin:
			add(in.Arg1)
			add(in.Arg2)
		case OpUnary, OpBranch:
			add(in.Arg1)
		case OpLoad:
			add(in.Arg1)
			add(in.Arg2)
		case OpStore:
			add(in.Arg1)
			add(in.Arg2)
			add(in.Dest)
		case OpCall:
			for _, a := range in.Args {
				add(a)
			}
		case OpReturn:
			for _, a := range in.Args {
				add(a)
			}
		}
	}
}

func (g *generator) stmt(s ast.Statement) error {
	if s == nil {
		return nil
	}
	switch x := s.(type) {
	case ast.Block:
		for _, s := range x.Statements {
			if e := g.stmt(s); e != nil {
				return e
			}
		}
	case *ast.Block:
		if x != nil {
			return g.stmt(*x)
		}
	case ast.Assignment:
		return g.assign(&x)
	case *ast.Assignment:
		return g.assign(x)
	case ast.If:
		return g.ifStmt(&x)
	case *ast.If:
		return g.ifStmt(x)
	case ast.While:
		return g.whileStmt(&x)
	case *ast.While:
		return g.whileStmt(x)
	case ast.Assert, *ast.Assert, ast.Assume, *ast.Assume:
	default:
		return fmt.Errorf("unsupported statement %T", s)
	}
	return nil
}
func (g *generator) assign(a *ast.Assignment) error {
	if len(a.Value) == 1 && len(a.Variables) > 1 {
		var call *ast.CallExpr
		switch c := a.Value[0].(type) {
		case *ast.CallExpr:
			call = c
		case ast.CallExpr:
			call = &c
		}
		if call != nil {
			c := call
			args, e := g.exprs(c.Args)
			if e != nil {
				return e
			}
			ds := make([]string, 0, len(a.Variables))
			for _, v := range a.Variables {
				n, ok := v.(ast.VariableExpr)
				if !ok {
					if p, yes := v.(*ast.VariableExpr); yes {
						ds = append(ds, p.Name)
						continue
					}
					return fmt.Errorf("multi-return target must be variable")
				}
				ds = append(ds, n.Name)
			}
			g.emit(Instr{Op: OpCall, Dests: ds, Arg1: c.Name, Args: args})
			return nil
		}
	}
	if len(a.Variables) != len(a.Value) {
		return fmt.Errorf("assignment target/value count mismatch: %d/%d", len(a.Variables), len(a.Value))
	}
	for i, v := range a.Value {
		val, e := g.expr(v)
		if e != nil {
			return e
		}
		switch t := a.Variables[i].(type) {
		case ast.VariableExpr:
			g.emit(Instr{Op: OpMove, Dest: t.Name, Arg1: val})
		case *ast.VariableExpr:
			g.emit(Instr{Op: OpMove, Dest: t.Name, Arg1: val})
		case ast.ArrayAccessExpr:
			idx, e := g.expr(t.Index)
			if e != nil {
				return e
			}
			g.emit(Instr{Op: OpStore, Dest: val, Arg1: t.Array, Arg2: idx})
		case *ast.ArrayAccessExpr:
			idx, e := g.expr(t.Index)
			if e != nil {
				return e
			}
			g.emit(Instr{Op: OpStore, Dest: val, Arg1: t.Array, Arg2: idx})
		default:
			return fmt.Errorf("unsupported assignment target %T", a.Variables[i])
		}
	}
	return nil
}
func (g *generator) ifStmt(x *ast.If) error {
	cond, err := g.condition(x.Condition)
	if err != nil {
		return err
	}
	thenLabel, endLabel := g.lab(), g.lab()
	if x.HasElse {
		elseLabel := endLabel
		endLabel = g.lab()
		g.emit(Instr{Op: OpBranch, Arg1: cond, Target: thenLabel})
		g.emit(Instr{Op: OpJump, Target: elseLabel})
		g.emit(Instr{Op: OpLabel, Target: thenLabel})
		if err = g.stmt(x.Then); err != nil {
			return err
		}
		g.emit(Instr{Op: OpJump, Target: endLabel})
		g.emit(Instr{Op: OpLabel, Target: elseLabel})
		if err = g.stmt(x.Else); err != nil {
			return err
		}
		g.emit(Instr{Op: OpLabel, Target: endLabel})
		return nil
	}
	g.emit(Instr{Op: OpBranch, Arg1: cond, Target: thenLabel})
	g.emit(Instr{Op: OpJump, Target: endLabel})
	g.emit(Instr{Op: OpLabel, Target: thenLabel})
	if err = g.stmt(x.Then); err != nil {
		return err
	}
	g.emit(Instr{Op: OpLabel, Target: endLabel})
	return nil
}
func (g *generator) whileStmt(x *ast.While) error {
	head, body, done := g.lab(), g.lab(), g.lab()
	g.emit(Instr{Op: OpLabel, Target: head})
	cond, err := g.condition(x.Condition)
	if err != nil {
		return err
	}
	g.emit(Instr{Op: OpBranch, Arg1: cond, Target: body})
	g.emit(Instr{Op: OpJump, Target: done})
	g.emit(Instr{Op: OpLabel, Target: body})
	if err = g.stmt(x.Body); err != nil {
		return err
	}
	g.emit(Instr{Op: OpJump, Target: head})
	g.emit(Instr{Op: OpLabel, Target: done})
	return nil
}
func (g *generator) expr(e ast.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("nil expression")
	}
	switch x := e.(type) {
	case ast.NumberExpr:
		if _, err := strconv.ParseInt(x.Value, 10, 64); err != nil {
			return "", err
		}
		t := g.tmp()
		g.emit(Instr{Op: OpConst, Dest: t, Arg1: x.Value})
		return t, nil
	case *ast.NumberExpr:
		return g.expr(*x)
	case ast.VariableExpr:
		return x.Name, nil
	case *ast.VariableExpr:
		return x.Name, nil
	case ast.UnaryExpr:
		v, e := g.expr(x.X)
		if e != nil {
			return "", e
		}
		t := g.tmp()
		g.emit(Instr{Op: OpUnary, Dest: t, Arg1: v, Operator: x.Op})
		return t, nil
	case *ast.UnaryExpr:
		return g.expr(*x)
	case ast.BinaryExpr:
		l, e := g.expr(x.Left)
		if e != nil {
			return "", e
		}
		r, e := g.expr(x.Right)
		if e != nil {
			return "", e
		}
		t := g.tmp()
		g.emit(Instr{Op: OpBin, Dest: t, Arg1: l, Arg2: r, Operator: x.Op})
		return t, nil
	case *ast.BinaryExpr:
		return g.expr(*x)
	case ast.ArrayAccessExpr:
		i, e := g.expr(x.Index)
		if e != nil {
			return "", e
		}
		t := g.tmp()
		g.emit(Instr{Op: OpLoad, Dest: t, Arg1: x.Array, Arg2: i})
		return t, nil
	case *ast.ArrayAccessExpr:
		return g.expr(*x)
	case ast.CallExpr:
		a, e := g.exprs(x.Args)
		if e != nil {
			return "", e
		}
		t := g.tmp()
		g.emit(Instr{Op: OpCall, Dest: t, Arg1: x.Name, Args: a})
		return t, nil
	case *ast.CallExpr:
		return g.expr(*x)
	default:
		return "", fmt.Errorf("unsupported expression %T", e)
	}
}
func (g *generator) exprs(es []ast.Expr) ([]string, error) {
	a := make([]string, 0, len(es))
	for _, e := range es {
		v, err := g.expr(e)
		if err != nil {
			return nil, err
		}
		a = append(a, v)
	}
	return a, nil
}
func (g *generator) condition(c ast.Condition) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil condition")
	}
	switch x := c.(type) {
	case ast.BoolCondition:
		v := "0"
		if x.Value {
			v = "1"
		}
		t := g.tmp()
		g.emit(Instr{Op: OpConst, Dest: t, Arg1: v})
		return t, nil
	case *ast.BoolCondition:
		return g.condition(*x)
	case ast.Comparison:
		l, e := g.expr(x.Left)
		if e != nil {
			return "", e
		}
		r, e := g.expr(x.Right)
		if e != nil {
			return "", e
		}
		t := g.tmp()
		g.emit(Instr{Op: OpBin, Dest: t, Arg1: l, Arg2: r, Operator: x.Op})
		return t, nil
	case *ast.Comparison:
		return g.condition(*x)
	case ast.NotCondition:
		v, e := g.condition(x.X)
		if e != nil {
			return "", e
		}
		t := g.tmp()
		g.emit(Instr{Op: OpUnary, Dest: t, Arg1: v, Operator: "!"})
		return t, nil
	case *ast.NotCondition:
		return g.condition(*x)
	case ast.BinaryCondition:
		l, e := g.condition(x.Left)
		if e != nil {
			return "", e
		}
		r, e := g.condition(x.Right)
		if e != nil {
			return "", e
		}
		t := g.tmp()
		g.emit(Instr{Op: OpBin, Dest: t, Arg1: l, Arg2: r, Operator: x.Op})
		return t, nil
	case *ast.BinaryCondition:
		return g.condition(*x)
	default:
		return "", fmt.Errorf("unsupported condition %T", c)
	}
}
