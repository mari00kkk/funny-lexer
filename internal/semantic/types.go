package semantic

import "funny-lexer/internal/ast"

type Type string

const (
	Unknown  Type = "<unknown>"
	Invalid  Type = "<invalid>"
	Int      Type = "int"
	IntArray Type = "int[]"
	Bool     Type = "bool"
)

func fromAST(t ast.Type) Type {
	switch t {
	case ast.IntType:
		return Int
	case ast.IntArray:
		return IntArray
	case "":
		return Unknown
	default:
		return Type(t)
	}
}
func compatible(a, b Type) bool {
	return a == Unknown || b == Unknown || a == Invalid || b == Invalid || a == b
}
