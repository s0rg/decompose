package cluster

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

type (
	ruleJSON struct {
		Name   string `json:"name"`
		Expr   string `json:"if"`
		Weight int    `json:"weight"`
	}

	rulePROG struct {
		Prog   *vm.Program
		Name   string
		Weight int
	}

	ruleENV struct {
		Node *node.View `expr:"node"`
	}

	exprRUN func(*vm.Program, any) (any, error)

	Rules struct {
		builder graph.NamedBuilderWriter
		runner  exprRUN
		nodes   map[string]*node.Node
		cluster map[string]map[string]*node.Ports
		rules   []*rulePROG
	}
)

func NewRules(
	b graph.NamedBuilderWriter,
	r exprRUN,
) *Rules {
	if r == nil {
		r = expr.Run
	}

	return &Rules{
		builder: b,
		runner:  r,
		nodes:   make(map[string]*node.Node),
		cluster: make(map[string]map[string]*node.Ports),
	}
}

func (cb *Rules) Name() string {
	return cb.builder.Name() + " clustered"
}

func (cb *Rules) Write(w io.Writer) error {
	for src, dmap := range cb.cluster {
		for dst, ports := range dmap {
			ports.Iter(func(_ string, plist []*node.Port) {
				for _, p := range plist {
					cb.builder.AddEdge(&node.Edge{
						SrcID:   src,
						DstID:   dst,
						SrcName: clusterPorts,
						DstName: clusterPorts,
						Port:    p,
					})
				}
			})
		}
	}

	if err := cb.builder.Write(w); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (cb *Rules) AddNode(n *node.Node) error {
	if cluster, ok := cb.Match(n); ok {
		n.Cluster = cluster
	}

	cb.nodes[n.ID] = n

	if err := cb.builder.AddNode(n); err != nil {
		return fmt.Errorf("builder: %w", err)
	}

	return nil
}

func (cb *Rules) AddEdge(e *node.Edge) {
	nsrc, ok := cb.nodes[e.SrcID]
	if !ok {
		return
	}

	ndst, ok := cb.nodes[e.DstID]
	if !ok {
		return
	}

	if nsrc.Cluster != ndst.Cluster {
		cdst, ok := cb.cluster[nsrc.Cluster]
		if !ok {
			cdst = make(map[string]*node.Ports)
		}

		var ports *node.Ports

		if ports, ok = cdst[ndst.Cluster]; !ok {
			ports = &node.Ports{}
			cdst[ndst.Cluster] = ports
		}

		ports.Add(clusterPorts, e.Port)

		cb.cluster[nsrc.Cluster] = cdst
	}

	cb.builder.AddEdge(e)
}

func (cb *Rules) CountRules() int {
	return len(cb.rules)
}

func (cb *Rules) FromReader(r io.Reader) (err error) {
	var rules []ruleJSON

	dec := json.NewDecoder(r)

	for dec.More() {
		if err = dec.Decode(&rules); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}

	opts := []expr.Option{
		expr.Env(ruleENV{}),
		expr.Optimize(true),
		expr.AsBool(),
	}

	for i := 0; i < len(rules); i++ {
		rule := &rules[i]

		prog, cerr := expr.Compile(rule.Expr, opts...)
		if cerr != nil {
			return fmt.Errorf("compile '%s': %w", rule.Expr, cerr)
		}

		cb.rules = append(cb.rules, &rulePROG{
			Name:   rule.Name,
			Weight: max(rule.Weight, 1),
			Prog:   prog,
		})
	}

	slices.SortFunc(cb.rules, func(a, b *rulePROG) int {
		return cmp.Compare(b.Weight, a.Weight)
	})

	return nil
}

func (cb *Rules) Match(n *node.Node) (cluster string, ok bool) {
	if len(cb.rules) == 0 {
		return
	}

	for _, rule := range cb.rules {
		res, err := cb.runner(rule.Prog, ruleENV{Node: n.ToView()})
		if err != nil {
			continue
		}

		resb, ok := res.(bool)
		if !ok || !resb {
			continue
		}

		return rule.Name, true
	}

	return "", false
}
