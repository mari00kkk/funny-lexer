package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"funny-lexer/internal/dfa"
	"funny-lexer/internal/ir"
	"funny-lexer/internal/parser"
	"funny-lexer/internal/runner"
	"funny-lexer/internal/semantic"
	"funny-lexer/internal/spec"
	"os"
	"path/filepath"
)

func write(path string, v string) error {
	if path == "" {
		return nil
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	return os.WriteFile(path, []byte(v), 0644)
}
func main() {
	input := flag.String("in", "", "input .funny file")
	out := flag.String("out", "", "optimized IR text output (default stdout)")
	before := flag.String("before", "", "unoptimized IR text output")
	jsonOut := flag.String("json", "", "JSON containing before/after IR")
	cfgDir := flag.String("cfg", "", "directory for CFG DOT files")
	graphDir := flag.String("graph", "", "directory for interference graph DOT files")
	noOpt := flag.Bool("no-opt", false, "disable all optimizations")
	constants := flag.Bool("const", true, "constant propagation and folding")
	algebra := flag.Bool("algebra", true, "algebraic simplification")
	dceFlag := flag.Bool("dce", true, "local dead-code elimination")
	cseFlag := flag.Bool("cse", false, "common subexpression elimination within a block")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/ir -in file.fun [options]")
		os.Exit(2)
	}
	src, e := os.ReadFile(*input)
	fatal(e)
	nfa := spec.BuildFunnyNFA()
	alphabet := spec.GetASCIIAlphabet()
	machine := dfa.Minimize(dfa.Build(nfa, alphabet), alphabet)
	tokens := runner.NewLexer(machine).Tokenize(string(src))
	p := parser.New(tokens)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		for _, x := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%d:%d: %s\n", x.Position.Line, x.Position.Column, x.Message)
		}
		os.Exit(1)
	}
	sem := semantic.Analyze(&program)
	if len(sem.Errors) > 0 {
		for _, x := range sem.Errors {
			fmt.Fprintln(os.Stderr, x.Message)
		}
		os.Exit(1)
	}
	raw, e := ir.Generate(&program)
	fatal(e)
	copyBytes, e := json.Marshal(raw)
	fatal(e)
	var pre ir.Program
	fatal(json.Unmarshal(copyBytes, &pre))
	opts := ir.OptOptions{Constants: *constants, Algebra: *algebra, DCE: *dceFlag, CSE: *cseFlag}
	if *noOpt {
		opts = ir.OptOptions{}
	}
	ir.Optimize(raw, opts)
	if *before != "" {
		fatal(write(*before, pre.String()))
	}
	if *out != "" {
		fatal(write(*out, raw.String()))
	} else {
		fmt.Print(raw.String())
	}
	type metric struct {
		BeforeInstructions int `json:"beforeInstructions"`
		AfterInstructions  int `json:"afterInstructions"`
		BeforeTemporaries  int `json:"beforeTemporaries"`
		AfterTemporaries   int `json:"afterTemporaries"`
		Colors             int `json:"colors"`
	}
	metrics := map[string]metric{}
	for i := range raw.Functions {
		f := raw.Functions[i]
		preF := pre.Functions[i]
		cfg, e := ir.BuildCFG(f)
		fatal(e)
		live := ir.AnalyzeLiveness(cfg)
		g := ir.BuildInterference(f, cfg, live)
		colors := ir.GreedyColor(g)
		metrics[f.Name] = metric{len(preF.Instructions), len(f.Instructions), countTemps(preF), countTemps(f), ir.ColorCount(colors)}
		if *cfgDir != "" {
			fatal(write(filepath.Join(*cfgDir, f.Name+".dot"), cfg.DOT(f.Name)))
		}
		if *graphDir != "" {
			fatal(write(filepath.Join(*graphDir, f.Name+".dot"), g.DOT()))
		}
		fmt.Fprintf(os.Stderr, "[%s] instructions %d -> %d; temporaries %d -> %d; colors %d\n", f.Name, len(preF.Instructions), len(f.Instructions), countTemps(preF), countTemps(f), ir.ColorCount(colors))
	}
	if *jsonOut != "" {
		data, e := json.MarshalIndent(struct {
			Before  *ir.Program       `json:"before"`
			After   *ir.Program       `json:"after"`
			Metrics map[string]metric `json:"metrics"`
		}{&pre, raw, metrics}, "", "  ")
		fatal(e)
		fatal(write(*jsonOut, string(data)+"\n"))
	}
}
func countTemps(f ir.Function) int {
	m := map[string]bool{}
	for _, i := range f.Instructions {
		if len(i.Dest) > 0 && i.Dest[0] == '%' {
			m[i.Dest] = true
		}
		for _, u := range []string{i.Arg1, i.Arg2} {
			if len(u) > 0 && u[0] == '%' {
				m[u] = true
			}
		}
		for _, u := range i.Args {
			if len(u) > 0 && u[0] == '%' {
				m[u] = true
			}
		}
	}
	return len(m)
}
func fatal(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
