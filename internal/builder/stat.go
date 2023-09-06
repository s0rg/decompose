//go:build !test

package builder

import (
	"cmp"
	"fmt"
	"io"
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

type portStat struct {
	Port  string
	Count int
}

type Stat struct {
	conns map[string]set.Unordered[string]
	ports map[string]int
	nodes int
	edges int
	exts  int
}

func NewStat() *Stat {
	return &Stat{
		ports: make(map[string]int),
		conns: make(map[string]set.Unordered[string]),
	}
}

func (s *Stat) Name() string {
	return "graph-stats"
}

func (s *Stat) AddNode(n *node.Node) error {
	if n.IsExternal() {
		s.exts++

		return nil
	}

	s.nodes++

	for _, p := range n.Ports {
		s.ports[p.Label()]++
	}

	s.conns[n.ID] = make(set.Unordered[string])

	return nil
}

func (s *Stat) AddEdge(src, dst string, _ *node.Port) {
	if _, ok := s.conns[dst]; !ok {
		return
	}

	conn, ok := s.conns[src]
	if !ok {
		return
	}

	if conn.Has(dst) {
		return
	}

	s.edges++
}

func (s *Stat) Write(w io.Writer) {
	ports := make([]*portStat, 0, len(s.ports))

	for k, c := range s.ports {
		ports = append(ports, &portStat{Port: k, Count: c})
	}

	slices.SortStableFunc(ports, func(a, b *portStat) int {
		return cmp.Compare(a.Count, b.Count)
	})

	slices.Reverse(ports)

	fmt.Fprintln(w, "Total:")
	fmt.Fprintf(w, "\tNodes:\t%d\n", s.nodes)
	fmt.Fprintf(w, "\tEdges:\t%d\n", s.edges)
	fmt.Fprintf(w, "\tExternals:\t%d\n", s.exts)

	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Ports:")

	for _, p := range ports {
		fmt.Fprintf(w, "\t%s:\t%d\n", p.Port, p.Count)
	}
}
