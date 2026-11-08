package ir

import "fmt"

type Block struct {
	ID           int     `json:"id"`
	Label        string  `json:"label,omitempty"`
	Instructions []Instr `json:"instructions"`
	Succ         []int   `json:"successors"`
	Pred         []int   `json:"predecessors"`
}
type CFG struct {
	Blocks []Block `json:"blocks"`
	Entry  int     `json:"entry"`
}

func BuildCFG(f Function) (*CFG, error) {
	if len(f.Instructions) == 0 {
		return &CFG{Entry: -1}, nil
	}
	leaders := map[int]bool{0: true}
	labels := map[string]int{}
	for i, in := range f.Instructions {
		if in.Op == OpLabel {
			leaders[i] = true
			labels[in.Target] = i
		}
		if (in.Op == OpJump || in.Op == OpBranch) && i+1 < len(f.Instructions) {
			leaders[i+1] = true
		}
	}
	starts := make([]int, 0, len(leaders))
	for i := range leaders {
		starts = append(starts, i)
	}
	sortInts(starts)
	cfg := &CFG{Entry: 0}
	insBlock := make([]int, len(f.Instructions))
	for n, s := range starts {
		e := len(f.Instructions)
		if n+1 < len(starts) {
			e = starts[n+1]
		}
		b := Block{ID: n}
		for i := s; i < e; i++ {
			insBlock[i] = n
			b.Instructions = append(b.Instructions, f.Instructions[i])
			if f.Instructions[i].Op == OpLabel {
				b.Label = f.Instructions[i].Target
			}
		}
		cfg.Blocks = append(cfg.Blocks, b)
	}
	labelBlock := map[string]int{}
	for l, i := range labels {
		labelBlock[l] = insBlock[i]
	}
	for n, b := range cfg.Blocks {
		last := Instr{}
		if len(b.Instructions) > 0 {
			last = b.Instructions[len(b.Instructions)-1]
		}
		add := func(id int) {
			if id < 0 || id >= len(cfg.Blocks) {
				return
			}
			for _, v := range cfg.Blocks[n].Succ {
				if v == id {
					return
				}
			}
			cfg.Blocks[n].Succ = append(cfg.Blocks[n].Succ, id)
		}
		switch last.Op {
		case OpJump:
			id, ok := labelBlock[last.Target]
			if !ok {
				return nil, fmt.Errorf("unknown label %q", last.Target)
			}
			add(id)
		case OpBranch:
			id, ok := labelBlock[last.Target]
			if !ok {
				return nil, fmt.Errorf("unknown label %q", last.Target)
			}
			add(id)
			if n+1 < len(cfg.Blocks) {
				add(n + 1)
			}
		case OpReturn:
		default:
			if n+1 < len(cfg.Blocks) {
				add(n + 1)
			}
		}
	}
	for n, b := range cfg.Blocks {
		for _, s := range b.Succ {
			cfg.Blocks[s].Pred = append(cfg.Blocks[s].Pred, n)
		}
	}
	return cfg, nil
}
func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}
