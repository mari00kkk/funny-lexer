package ir

import (
	"fmt"
	"strings"
)

func (p *Program) String() string {
	var b strings.Builder
	for _, f := range p.Functions {
		fmt.Fprintf(&b, "function %s(%s) -> (%s) {\n", f.Name, strings.Join(f.Parameters, ", "), strings.Join(f.Returns, ", "))
		for _, i := range f.Instructions {
			fmt.Fprintf(&b, "  %s\n", i.String())
		}
		b.WriteString("}\n")
	}
	return b.String()
}
func (c *CFG) DOT(name string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "digraph %q {\n", name)
	for _, x := range c.Blocks {
		var lines []string
		for _, i := range x.Instructions {
			lines = append(lines, i.String())
		}
		fmt.Fprintf(&b, "  B%d [shape=box,label=\"B%d: %s\"];\n", x.ID, x.ID, strings.ReplaceAll(strings.Join(lines, "\\l"), "\"", "\\\""))
	}
	for _, x := range c.Blocks {
		for _, s := range x.Succ {
			fmt.Fprintf(&b, "  B%d -> B%d;\n", x.ID, s)
		}
	}
	b.WriteString("}\n")
	return b.String()
}
func (g Graph) DOT() string {
	var b strings.Builder
	b.WriteString("graph interference {\n")
	for n := range g {
		fmt.Fprintf(&b, "  \"%s\";\n", n)
		for v := range g[n] {
			if n < v {
				fmt.Fprintf(&b, "  \"%s\" -- \"%s\";\n", n, v)
			}
		}
	}
	b.WriteString("}\n")
	return b.String()
}
