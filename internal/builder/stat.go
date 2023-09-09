package builder

import (
	"cmp"
	"fmt"
	"io"
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

const minClusters = 2
const defaultName = "default"

type stat struct {
	Name  string
	Count int
}

type Stat struct {
	conns      map[string]set.Unordered[string]
	ports      map[string]int
	clusters   map[string]int
	nodes      int
	edgesUniq  int
	edgesTotal int
	externals  int
}

func NewStat() *Stat {
	return &Stat{
		ports:    make(map[string]int),
		clusters: make(map[string]int),
		conns:    make(map[string]set.Unordered[string]),
	}
}

func (s *Stat) Name() string {
	return "graph-stats"
}

func (s *Stat) AddNode(n *node.Node) error {
	if n.IsExternal() {
		s.externals++

		return nil
	}

	s.nodes++

	for _, p := range n.Ports {
		s.ports[p.Label()]++
	}

	s.conns[n.ID] = make(set.Unordered[string])

	cluster := n.Cluster
	if cluster == "" {
		cluster = defaultName
	}

	s.clusters[cluster]++

	return nil
}

func (s *Stat) isSuitable(srcID, dstID string) (uniq, yes bool) {
	sc, ok := s.conns[srcID]
	if !ok {
		return
	}

	dc, ok := s.conns[dstID]
	if !ok {
		return
	}

	uniq = !(sc.Has(dstID) || dc.Has(srcID))

	sc.Add(dstID)

	return uniq, true
}

func (s *Stat) AddEdge(srcID, dstID string, _ *node.Port) {
	uniq, ok := s.isSuitable(srcID, dstID)
	if !ok {
		return
	}

	if uniq {
		s.edgesUniq++
	}

	s.edgesTotal++
}

func (s *Stat) Write(w io.Writer) {
	fmt.Fprintf(w, "Nodes: %d\n", s.nodes)
	fmt.Fprintf(w, "Connections total: %d uniq: %d\n", s.edgesTotal, s.edgesUniq)

	if s.externals > 0 {
		fmt.Fprintf(w, "Externals: %d\n", s.externals)
	}

	fmt.Fprintln(w, "")

	ports, clusters := s.calcStats()

	if len(clusters) > 0 {
		fmt.Fprintln(w, "Clusters:")
		writeStats(w, clusters)
	}

	fmt.Fprintln(w, "Ports:")
	writeStats(w, ports)
}

func (s *Stat) calcStats() (ports, clusters []*stat) {
	ports = make([]*stat, 0, len(s.ports))

	for k, c := range s.ports {
		ports = append(ports, &stat{Name: k, Count: c})
	}

	slices.SortStableFunc(ports, byCount)
	slices.Reverse(ports)

	if len(s.clusters) < minClusters {
		return ports, clusters
	}

	clusters = make([]*stat, 0, len(s.clusters))

	for k, c := range s.clusters {
		clusters = append(clusters, &stat{Name: k, Count: c})
	}

	slices.SortStableFunc(clusters, byCount)
	slices.Reverse(clusters)

	return ports, clusters
}

func byCount(a, b *stat) int {
	return cmp.Compare(a.Count, b.Count)
}

func writeStats(w io.Writer, s []*stat) {
	for _, v := range s {
		fmt.Fprintf(w, "\t%s: %d\n", v.Name, v.Count)
	}

	fmt.Fprintln(w, "")
}