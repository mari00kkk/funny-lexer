package ir

type Graph map[string]map[string]bool

func BuildInterference(f Function, c *CFG, live LiveInfo) Graph {
	g := Graph{}
	add := func(a, b string) {
		if len(a) == 0 || len(b) == 0 || a[0] != '%' || b[0] != '%' || a == b {
			return
		}
		if g[a] == nil {
			g[a] = map[string]bool{}
		}
		if g[b] == nil {
			g[b] = map[string]bool{}
		}
		g[a][b] = true
		g[b][a] = true
	}
	for i, b := range c.Blocks {
		set := map[string]bool{}
		for v := range live.Out[i] {
			set[v] = true
		}
		for j := len(b.Instructions) - 1; j >= 0; j-- {
			in := b.Instructions[j]
			instructionDefs := defs(in)
			for _, d := range instructionDefs {
				if len(d) > 0 && d[0] == '%' && g[d] == nil {
					g[d] = map[string]bool{}
				}
				for v := range set {
					add(d, v)
				}
				for _, other := range instructionDefs {
					add(d, other)
				}
				delete(set, d)
			}
			for _, u := range uses(in) {
				set[u] = true
			}
		}
	}
	return g
}
