package ir

import "strconv"

type OptOptions struct{ Constants, Algebra, DCE, CSE bool }

func DefaultOptOptions() OptOptions { return OptOptions{Constants: true, Algebra: true, DCE: true} }
func Optimize(p *Program, o OptOptions) {
	if p == nil {
		return
	}
	for fi := range p.Functions {
		f := &p.Functions[fi]
		ranges := blockRanges(f.Instructions)
		rebuilt := make([]Instr, 0, len(f.Instructions))
		for _, r := range ranges {
			block := append([]Instr(nil), f.Instructions[r[0]:r[1]]...)
			if o.Constants || o.Algebra {
				propagateRange(block, o)
			}
			if o.CSE {
				cseRange(block)
			}
			if o.DCE {
				dceRange(&block, 0, len(block))
			}
			rebuilt = append(rebuilt, block...)
		}
		f.Instructions = rebuilt
	}
}
func blockRanges(a []Instr) [][2]int {
	if len(a) == 0 {
		return nil
	}
	leaders := map[int]bool{0: true}
	for i, x := range a {
		if x.Op == OpLabel {
			leaders[i] = true
		}
		if (x.Op == OpJump || x.Op == OpBranch || x.Op == OpReturn) && i+1 < len(a) {
			leaders[i+1] = true
		}
	}
	var starts []int
	for i := range leaders {
		starts = append(starts, i)
	}
	sortInts(starts)
	out := make([][2]int, 0, len(starts))
	for i, s := range starts {
		e := len(a)
		if i+1 < len(starts) {
			e = starts[i+1]
		}
		out = append(out, [2]int{s, e})
	}
	return out
}
func isInt(s string) (int64, bool) { v, e := strconv.ParseInt(s, 10, 64); return v, e == nil }
func replace(s string, m map[string]string) string {
	seen := map[string]bool{}
	for {
		v, ok := m[s]
		if !ok || seen[s] {
			return s
		}
		seen[s] = true
		s = v
	}
}
func eval(op string, a, b int64) (int64, bool) {
	switch op {
	case "+":
		return a + b, true
	case "-":
		return a - b, true
	case "*":
		return a * b, true
	case "/":
		if b != 0 {
			return a / b, true
		}
	case "<":
		if a < b {
			return 1, true
		}
		return 0, true
	case "<=":
		if a <= b {
			return 1, true
		}
		return 0, true
	case ">":
		if a > b {
			return 1, true
		}
		return 0, true
	case ">=":
		if a >= b {
			return 1, true
		}
		return 0, true
	case "==":
		if a == b {
			return 1, true
		}
		return 0, true
	case "!=":
		if a != b {
			return 1, true
		}
		return 0, true
	case "&&", "and":
		if a != 0 && b != 0 {
			return 1, true
		}
		return 0, true
	case "||", "or":
		if a != 0 || b != 0 {
			return 1, true
		}
		return 0, true
	case "->":
		if a == 0 || b != 0 {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}
func propagateRange(ins []Instr, o OptOptions) {
	consts := map[string]string{}
	for i := range ins {
		x := &ins[i]
		x.Arg1 = replace(x.Arg1, consts)
		x.Arg2 = replace(x.Arg2, consts)
		if x.Op != OpReturn {
			for j := range x.Args {
				x.Args[j] = replace(x.Args[j], consts)
			}
		}
		if x.Op == OpBin && o.Constants {
			a, oka := isInt(x.Arg1)
			b, okb := isInt(x.Arg2)
			if oka && okb {
				if v, ok := eval(x.Operator, a, b); ok {
					x.Op = OpConst
					x.Arg1 = strconv.FormatInt(v, 10)
					x.Arg2 = ""
					x.Operator = ""
				}
			}
		}
		if x.Op == OpUnary && o.Constants {
			if a, ok := isInt(x.Arg1); ok {
				switch x.Operator {
				case "-":
					x.Op = OpConst
					x.Arg1 = strconv.FormatInt(-a, 10)
					x.Operator = ""
				case "!":
					v := 0
					if a == 0 {
						v = 1
					}
					x.Op = OpConst
					x.Arg1 = strconv.Itoa(v)
					x.Operator = ""
				}
			}
		}
		if x.Op == OpBin && o.Algebra {
			switch {
			case (x.Operator == "+" || x.Operator == "-") && x.Arg2 == "0", (x.Operator == "*" || x.Operator == "/") && x.Arg2 == "1":
				x.Op = OpMove
				x.Arg2 = ""
				x.Operator = ""
			case (x.Operator == "+" || x.Operator == "*") && x.Arg1 == "0" && x.Operator == "+", x.Operator == "*" && x.Arg1 == "1":
				x.Op = OpMove
				x.Arg1 = x.Arg2
				x.Arg2 = ""
				x.Operator = ""
			case x.Operator == "*" && (x.Arg1 == "0" || x.Arg2 == "0"):
				x.Op = OpConst
				x.Arg1 = "0"
				x.Arg2 = ""
				x.Operator = ""
			}
		}
		for _, d := range defs(*x) {
			delete(consts, d)
			for name, value := range consts {
				if value == d {
					delete(consts, name)
				}
			}
			if x.Op == OpConst {
				consts[d] = x.Arg1
			} else if x.Op == OpMove {
				consts[d] = x.Arg1
			}
		}
		if x.Op == OpCall {
			consts = map[string]string{}
		}
	}
}
func uses(i Instr) []string {
	var u []string
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := isInt(s); ok {
			return
		}
		u = append(u, s)
	}
	switch i.Op {
	case OpMove, OpUnary, OpBranch, OpParam:
		add(i.Arg1)
	case OpBin:
		add(i.Arg1)
		add(i.Arg2)
	case OpLoad:
		add(i.Arg2)
	case OpStore:
		add(i.Dest)
		add(i.Arg2)
	case OpCall:
		for _, a := range i.Args {
			add(a)
		}
	case OpReturn:
		for _, a := range i.Args {
			add(a)
		}
	}
	return u
}
func defs(i Instr) []string {
	if i.Op == OpCall {
		if len(i.Dests) > 0 {
			return i.Dests
		}
		if i.Dest != "" && i.Dest[0] == '%' {
			return []string{i.Dest}
		}
		return nil
	}
	d := def(i)
	if d != "" {
		return []string{d}
	}
	return nil
}
func def(i Instr) string {
	switch i.Op {
	case OpConst, OpMove, OpBin, OpUnary, OpLoad:
		return i.Dest
	}
	return ""
}
func dceRange(ins *[]Instr, start, end int) {
	a := *ins
	if start >= end {
		return
	}
	live := map[string]bool{}
	for i := end - 1; i >= start; i-- {
		x := a[i]
		d := def(x)
		side := x.Op == OpCall || x.Op == OpLoad || x.Op == OpStore || x.Op == OpJump || x.Op == OpBranch || x.Op == OpLabel || x.Op == OpReturn || x.Op == OpParam
		if d != "" && len(d) > 0 && d[0] == '%' && !live[d] && !side {
			a = append(a[:i], a[i+1:]...)
			*ins = a
			continue
		}
		if d != "" {
			delete(live, d)
		}
		for _, u := range uses(x) {
			live[u] = true
		}
	}
}
func cseRange(ins []Instr) {
	table := map[string]string{}
	for i := range ins {
		x := &ins[i]
		if x.Op == OpCall || x.Op == OpStore || x.Op == OpLoad || ((x.Op == OpMove || x.Op == OpConst) && x.Dest != "" && x.Dest[0] != '%') {
			table = map[string]string{}
		}
		if x.Op == OpBin {
			key := x.Operator + "|" + x.Arg1 + "|" + x.Arg2
			if v, ok := table[key]; ok {
				x.Op = OpMove
				x.Arg1 = v
				x.Arg2 = ""
				x.Operator = ""
			} else {
				table[key] = x.Dest
			}
		}
		if d := def(*x); d != "" {
			for k, v := range table {
				if v == d {
					delete(table, k)
				}
			}
		}
	}
}
