package ir

type LiveInfo struct{ In, Out []map[string]bool }

func AnalyzeLiveness(c *CFG) LiveInfo {
	n := len(c.Blocks)
	r := LiveInfo{In: make([]map[string]bool, n), Out: make([]map[string]bool, n)}
	use := make([]map[string]bool, n)
	defsets := make([]map[string]bool, n)
	for i, b := range c.Blocks {
		use[i] = map[string]bool{}
		defsets[i] = map[string]bool{}
		for _, in := range b.Instructions {
			for _, u := range uses(in) {
				if !defsets[i][u] {
					use[i][u] = true
				}
			}
			for _, d := range defs(in) {
				defsets[i][d] = true
			}
		}
		r.In[i] = map[string]bool{}
		r.Out[i] = map[string]bool{}
	}
	changed := true
	for changed {
		changed = false
		for i := n - 1; i >= 0; i-- {
			out := map[string]bool{}
			for _, s := range c.Blocks[i].Succ {
				for v := range r.In[s] {
					out[v] = true
				}
			}
			in := map[string]bool{}
			for v := range use[i] {
				in[v] = true
			}
			for v := range out {
				if !defsets[i][v] {
					in[v] = true
				}
			}
			if !sameSet(out, r.Out[i]) || !sameSet(in, r.In[i]) {
				r.Out[i] = out
				r.In[i] = in
				changed = true
			}
		}
	}
	return r
}
func sameSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
