package semantic

import (
	"encoding/json"
	"fmt"
	"funny-lexer/internal/ast"
	"os"
)

type Result struct {
	Errors  []Error           `json:"errors"`
	Types   map[ast.Expr]Type `json:"-"`
	Symbols map[string]Symbol `json:"symbols"`
}
type Analyzer struct {
	result        *Result
	global, scope *Scope
}

func New() *Analyzer {
	g := newScope(nil)
	return &Analyzer{result: &Result{Types: map[ast.Expr]Type{}, Symbols: map[string]Symbol{}}, global: g, scope: g}
}
func Analyze(p *ast.Program) *Result { return New().Analyze(p) }
func (a *Analyzer) add(msg string)   { a.result.Errors = append(a.result.Errors, Error{Message: msg}) }
func (a *Analyzer) Analyze(p *ast.Program) *Result {
	if p == nil {
		a.add("nil program")
		return a.result
	}
	// Register signatures before checking bodies so forward calls work.
	for i := range p.Functions {
		f := &p.Functions[i]
		ps := make([]Type, len(f.Parameters))
		rs := make([]Type, len(f.Returns))
		for j, v := range f.Parameters {
			ps[j] = fromAST(v.Type)
		}
		for j, v := range f.Returns {
			rs[j] = fromAST(v.Type)
		}
		if !a.global.declare(Symbol{Name: f.Name, Kind: FunctionSymbol, Parameters: ps, Returns: rs}) {
			a.add(fmt.Sprintf("duplicate declaration of %q", f.Name))
		}
	}
	for i := range p.Formulas {
		f := &p.Formulas[i]
		ps := make([]Type, len(f.Parameters))
		for j, v := range f.Parameters {
			ps[j] = fromAST(v.Type)
		}
		if !a.global.declare(Symbol{Name: f.Name, Kind: FormulaSymbol, Type: Bool, Parameters: ps, Returns: []Type{Bool}}) {
			a.add(fmt.Sprintf("duplicate declaration of %q", f.Name))
		}
	}
	for i := range p.Functions {
		a.checkFunction(&p.Functions[i])
	}
	for i := range p.Formulas {
		a.checkFormula(&p.Formulas[i])
	}
	for n, s := range a.global.Symbols {
		a.result.Symbols[n] = s
	}
	return a.result
}
func (a *Analyzer) declare(v ast.Variable, role string) {
	t := fromAST(v.Type)
	if v.Name == "" {
		return
	}
	if !a.scope.declare(Symbol{Name: v.Name, Kind: VariableSymbol, Type: t, Inferred: t == Unknown}) {
		a.add(fmt.Sprintf("duplicate declaration of %q (%s)", v.Name, role))
	}
}
func (a *Analyzer) checkFunction(f *ast.Function) {
	old := a.scope
	a.scope = newScope(a.global)
	for _, v := range f.Parameters {
		a.declare(v, "parameter")
	}
	for _, v := range f.Returns {
		a.declare(v, "return variable")
	}
	for _, v := range f.Uses {
		a.declare(v, "local variable")
	}
	if f.Requires != nil {
		a.checkPredicate(*f.Requires)
	}
	if f.Body != nil {
		a.checkStatement(f.Body)
	}
	if f.Ensures != nil {
		a.checkPredicate(*f.Ensures)
	}
	a.scope = old
}
func (a *Analyzer) checkFormula(f *ast.Formula) {
	old := a.scope
	a.scope = newScope(a.global)
	for _, v := range f.Parameters {
		a.declare(v, "formula parameter")
	}
	if f.Body != nil {
		a.checkPredicate(f.Body)
	}
	a.scope = old
}
func (r *Result) WriteJSON(path string) error {
	type ann struct {
		Node string `json:"node"`
		Type Type   `json:"type"`
	}
	out := struct {
		Annotations []ann             `json:"annotations"`
		Errors      []Error           `json:"errors"`
		Symbols     map[string]Symbol `json:"symbols"`
	}{Errors: r.Errors, Symbols: r.Symbols}
	for e, t := range r.Types {
		out.Annotations = append(out.Annotations, ann{fmt.Sprintf("%T", e), t})
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
