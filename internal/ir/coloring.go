package ir

import "sort"

func GreedyColor(g Graph) map[string]int {
	nodes := make([]string, 0, len(g))
	for n := range g {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		di, dj := len(g[nodes[i]]), len(g[nodes[j]])
		if di == dj {
			return nodes[i] < nodes[j]
		}
		return di > dj
	})
	colors := map[string]int{}
	for _, n := range nodes {
		used := map[int]bool{}
		for v := range g[n] {
			if c, ok := colors[v]; ok {
				used[c] = true
			}
		}
		c := 0
		for used[c] {
			c++
		}
		colors[n] = c
	}
	return colors
}
func ColorCount(c map[string]int) int {
	m := -1
	for _, v := range c {
		if v > m {
			m = v
		}
	}
	return m + 1
}
