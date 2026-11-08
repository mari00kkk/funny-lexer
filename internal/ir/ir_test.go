package ir

import (
	"funny-lexer/internal/ast"
	"strings"
	"testing"
)

func TestGenerateIfAndWhileCFG(t *testing.T) {
	p := &ast.Program{Functions: []ast.Function{{Name: "f", Returns: []ast.Variable{{Name: "r", Type: ast.IntType}}, Body: ast.Block{Statements: []ast.Statement{
		ast.If{Condition: ast.Comparison{Left: ast.VariableExpr{Name: "x"}, Op: ">", Right: ast.NumberExpr{Value: "0"}}, Then: ast.Assignment{Variables: []ast.Expr{ast.VariableExpr{Name: "r"}}, Value: []ast.Expr{ast.NumberExpr{Value: "1"}}}, Else: ast.Assignment{Variables: []ast.Expr{ast.VariableExpr{Name: "r"}}, Value: []ast.Expr{ast.NumberExpr{Value: "2"}}}, HasElse: true},
		ast.While{Condition: ast.BoolCondition{Value: false}, Body: ast.Assignment{Variables: []ast.Expr{ast.VariableExpr{Name: "r"}}, Value: []ast.Expr{ast.NumberExpr{Value: "3"}}}},
	}}}}}
	got, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	f := got.Functions[0]
	cfg, err := BuildCFG(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Blocks) < 5 {
		t.Fatalf("expected split CFG blocks, got %d", len(cfg.Blocks))
	}

	var branch Instr
	for _, in := range f.Instructions {
		if in.Op == OpBranch {
			branch = in
			break
		}
	}
	if branch.Target == "" {
		t.Fatal("missing conditional branch")
	}
	targetFound := false
	for _, in := range f.Instructions {
		if in.Op == OpLabel && in.Target == branch.Target {
			targetFound = true
		}
	}
	if !targetFound {
		t.Fatalf("branch target %q not defined", branch.Target)
	}
}
func TestLocalOptimizations(t *testing.T) {
	f := Function{Name: "f", Instructions: []Instr{
		{Op: OpConst, Dest: "%t1", Arg1: "2"},
		{Op: OpConst, Dest: "%t2", Arg1: "3"},
		{Op: OpBin, Dest: "%t3", Arg1: "%t1", Arg2: "%t2", Operator: "+"},
		{Op: OpBin, Dest: "%t4", Arg1: "%t1", Arg2: "%t2", Operator: "+"},
		{Op: OpMove, Dest: "result", Arg1: "%t4"},
		{Op: OpReturn, Args: []string{"result"}},
	}}
	p := &Program{Functions: []Function{f}}
	Optimize(p, OptOptions{Constants: true, Algebra: true, DCE: true, CSE: true})
	out := p.Functions[0].Instructions
	if len(out) != 2 {
		t.Fatalf("expected 3 instructions after folding/CSE/DCE, got %d: %s", len(out), p.String())
	}
	if out[0].Op != OpMove || out[0].Dest != "result" || out[0].Arg1 != "5" {
		t.Fatalf("expected result = 5 after folding, got %+v", out[0])
	}
	if !strings.Contains(p.String(), "result = 5") {
		t.Fatalf("unexpected optimized IR: %s", p.String())
	}
}
func TestInterferenceAndColoring(t *testing.T) {
	f := Function{Name: "f", Instructions: []Instr{{Op: OpConst, Dest: "%t1", Arg1: "1"}, {Op: OpConst, Dest: "%t2", Arg1: "2"}, {Op: OpBin, Dest: "%t3", Arg1: "%t1", Arg2: "%t2", Operator: "+"}, {Op: OpReturn, Args: []string{"%t3"}}}}
	cfg, err := BuildCFG(f)
	if err != nil {
		t.Fatal(err)
	}
	live := AnalyzeLiveness(cfg)
	g := BuildInterference(f, cfg, live)
	if !g["%t1"]["%t2"] {
		t.Fatalf("expected simultaneously live temporaries to interfere: %#v", g)
	}
	colors := GreedyColor(g)
	if ColorCount(colors) < 2 {
		t.Fatalf("expected at least two colors, got %v", colors)
	}
}

func TestConstantFactsDoNotCrossBasicBlocks(t *testing.T) {
	f := Function{Name: "f", Instructions: []Instr{
		{Op: OpConst, Dest: "%t1", Arg1: "1"},
		{Op: OpBranch, Arg1: "%t1", Target: "L1"},
		{Op: OpLabel, Target: "L1"},
		{Op: OpMove, Dest: "result", Arg1: "%t1"},
		{Op: OpReturn, Args: []string{"result"}},
	}}
	p := &Program{Functions: []Function{f}}
	Optimize(p, OptOptions{Constants: true})

	for _, in := range p.Functions[0].Instructions {
		if in.Op == OpMove && in.Dest == "result" && in.Arg1 == "1" {
			t.Fatal("constant propagated across basic-block boundary")
		}
	}
}

func TestCopyPropagationAndSourceVariableRedefinition(t *testing.T) {
	f := Function{Name: "f", Instructions: []Instr{
		{Op: OpMove, Dest: "%t1", Arg1: "x"},
		{Op: OpMove, Dest: "x", Arg1: "y"},
		{Op: OpMove, Dest: "y", Arg1: "z"},
		{Op: OpMove, Dest: "result", Arg1: "%t1"},
		{Op: OpReturn, Args: []string{"result"}},
	}}
	p := &Program{Functions: []Function{f}}
	Optimize(p, OptOptions{Constants: true, Algebra: true})

	for _, in := range p.Functions[0].Instructions {
		if in.Dest == "result" && in.Arg1 == "z" {
			t.Fatal("stale copy propagated through redefinition")
		}
	}
}

func TestGenerateMultiResultCall(t *testing.T) {
	p := &ast.Program{Functions: []ast.Function{{Name: "caller", Returns: []ast.Variable{{Name: "left", Type: ast.IntType}, {Name: "right", Type: ast.IntType}}, Body: ast.Assignment{Variables: []ast.Expr{ast.VariableExpr{Name: "left"}, ast.VariableExpr{Name: "right"}}, Value: []ast.Expr{ast.CallExpr{Name: "pair"}}}}}}
	got, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	ins := got.Functions[0].Instructions
	if len(ins) < 1 || ins[0].Op != OpCall || len(ins[0].Dests) != 2 || ins[0].Dests[0] != "left" || ins[0].Dests[1] != "right" {
		t.Fatalf("expected multi-result call destinations, got %+v", ins)
	}
}
