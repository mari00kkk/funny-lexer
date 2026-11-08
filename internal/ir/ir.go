package ir

import (
	"fmt"
	"strings"
)

type Op string

const (
	OpConst  Op = "const"
	OpMove   Op = "move"
	OpBin    Op = "bin"
	OpUnary  Op = "unary"
	OpLoad   Op = "load"
	OpStore  Op = "store"
	OpCall   Op = "call"
	OpParam  Op = "param"
	OpReturn Op = "return"
	OpLabel  Op = "label"
	OpJump   Op = "jump"
	OpBranch Op = "branch"
	OpNop    Op = "nop"
)

type Instr struct {
	Op       Op       `json:"op"`
	Dest     string   `json:"dest,omitempty"`
	Dests    []string `json:"dests,omitempty"`
	Arg1     string   `json:"arg1,omitempty"`
	Arg2     string   `json:"arg2,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Args     []string `json:"args,omitempty"`
	Target   string   `json:"target,omitempty"`
	Comment  string   `json:"comment,omitempty"`
}

func (i Instr) String() string {
	switch i.Op {
	case OpConst:
		return fmt.Sprintf("%s = %s", i.Dest, i.Arg1)
	case OpMove:
		return fmt.Sprintf("%s = %s", i.Dest, i.Arg1)
	case OpBin:
		return fmt.Sprintf("%s = %s %s %s", i.Dest, i.Arg1, i.Operator, i.Arg2)
	case OpUnary:
		return fmt.Sprintf("%s = %s%s", i.Dest, i.Operator, i.Arg1)
	case OpLoad:
		return fmt.Sprintf("%s = %s[%s]", i.Dest, i.Arg1, i.Arg2)
	case OpStore:
		return fmt.Sprintf("%s[%s] = %s", i.Arg1, i.Arg2, i.Dest)
	case OpCall:
		dest := i.Dest
		if len(i.Dests) > 0 {
			dest = strings.Join(i.Dests, ", ")
		}
		if dest != "" {
			return fmt.Sprintf("%s = call %s(%v)", dest, i.Arg1, i.Args)
		}
		return fmt.Sprintf("call %s(%v)", i.Arg1, i.Args)
	case OpParam:
		return "param " + i.Arg1
	case OpReturn:
		return fmt.Sprintf("return %v", i.Args)
	case OpLabel:
		return i.Target + ":"
	case OpJump:
		return "goto " + i.Target
	case OpBranch:
		return fmt.Sprintf("if %s != 0 goto %s", i.Arg1, i.Target)
	default:
		return "nop"
	}
}

type Function struct {
	Name           string            `json:"name"`
	Parameters     []string          `json:"parameters,omitempty"`
	ParameterTypes []string          `json:"parameterTypes,omitempty"`
	Returns        []string          `json:"returns,omitempty"`
	ReturnTypes    []string          `json:"returnTypes,omitempty"`
	Locals         []string          `json:"locals,omitempty"`
	LocalTypes     map[string]string `json:"localTypes,omitempty"`
	Instructions   []Instr           `json:"instructions"`
}
type Program struct {
	Functions []Function `json:"functions"`
}
