package semantic

type SymbolKind string

const (
	VariableSymbol SymbolKind = "variable"
	FunctionSymbol SymbolKind = "function"
	FormulaSymbol  SymbolKind = "formula"
)

type Symbol struct {
	Name       string     `json:"name"`
	Kind       SymbolKind `json:"kind"`
	Type       Type       `json:"type,omitempty"`
	Parameters []Type     `json:"parameters,omitempty"`
	Returns    []Type     `json:"returns,omitempty"`
	Inferred   bool       `json:"inferred,omitempty"`
}
type Scope struct {
	Parent  *Scope
	Symbols map[string]Symbol
}

func newScope(parent *Scope) *Scope { return &Scope{Parent: parent, Symbols: map[string]Symbol{}} }
func (s *Scope) lookup(name string) (Symbol, bool) {
	for p := s; p != nil; p = p.Parent {
		if v, ok := p.Symbols[name]; ok {
			return v, true
		}
	}
	return Symbol{}, false
}
func (s *Scope) declare(v Symbol) bool {
	if _, ok := s.Symbols[v.Name]; ok {
		return false
	}
	s.Symbols[v.Name] = v
	return true
}
