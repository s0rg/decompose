package cluster

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/s0rg/set"
	"github.com/s0rg/trie"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

type Layers struct {
	edges      map[string]map[string]node.Ports
	nodes      map[string]*node.Node
	remotes    set.Unordered[string]
	b          graph.NamedBuilderWriter
	g          connGraph
	similarity float64
}

func NewLayers(
	b graph.NamedBuilderWriter,
	s float64,
) *Layers {
	return &Layers{
		b:          NewRules(b, nil),
		g:          make(connGraph),
		edges:      make(map[string]map[string]node.Ports),
		nodes:      make(map[string]*node.Node),
		remotes:    make(set.Unordered[string]),
		similarity: s,
	}
}

func (l *Layers) Name() string {
	return l.b.Name() + " auto:" +
		strconv.FormatFloat(l.similarity, 'f', 1, 64)
}

func (l *Layers) AddNode(n *node.Node) error {
	l.nodes[n.ID] = n

	if n.IsExternal() {
		l.remotes.Add(n.ID)

		return nil
	}

	l.g.AddNode(n)

	return nil
}

func (l *Layers) upsertEdge(src, dst string, p *node.Port) (rv node.Ports) {
	dest, ok := l.edges[src]
	if !ok {
		dest = make(map[string]node.Ports)
	}

	if rv, ok = dest[dst]; !ok {
		rv = make(node.Ports, 0, 1)
	}

	rv = append(rv, p)

	dest[dst] = rv
	l.edges[src] = dest

	return rv
}

func (l *Layers) AddEdge(src, dst string, p *node.Port) {
	l.upsertEdge(src, dst, p)

	if l.remotes.Has(src) || l.remotes.Has(dst) {
		return
	}

	l.g.AddEdge(src, dst)
}

func (l *Layers) names(ids []string) (rv []string) {
	rv = make([]string, 0, len(ids))

	for _, id := range ids {
		n := l.nodes[id]

		rv = append(rv, n.Name)
	}

	return rv
}

func (l *Layers) Write(w io.Writer) error {
	var (
		seen  = make(set.Unordered[string])
		layer []string
	)

	const maxLabelParts = 3

	for i := 0; ; i++ {
		layer = l.g.NextLayer(layer, seen)
		if len(layer) == 0 {
			break
		}

		grp := NewGrouper(l.similarity)

		for _, id := range layer {
			grp.Add(id, l.g[id])
		}

		grp.IterGroups(func(id int, membersID []string) {
			label := CreateLabel(l.names(membersID), maxLabelParts)

			for _, mid := range membersID {
				n := l.nodes[mid]
				n.Cluster = fmt.Sprintf("l%02d-%02d-%s", i, id, label)

				_ = l.b.AddNode(n)

				delete(l.nodes, mid)
			}
		})
	}

	// store remains
	for _, n := range l.nodes {
		_ = l.b.AddNode(n)
	}

	for src, dmap := range l.edges {
		for dst, ports := range dmap {
			for _, p := range ports {
				l.b.AddEdge(src, dst, p)
			}
		}
	}

	if err := l.b.Write(w); err != nil {
		return fmt.Errorf("auto: %w", err)
	}

	return nil
}

func CreateLabel(names []string, nmax int) (rv string) {
	t := trie.New[struct{}]()
	v := struct{}{}

	for _, n := range names {
		t.Add(n, v)
	}

	const (
		root    = ""
		minus   = "-"
		cutset  = "1234567890" + minus
		maxdiff = 3
	)

	comm := slices.DeleteFunc(t.Common(root, maxdiff), func(v string) bool {
		s, _ := t.Suggest(v)

		return len(s) == 1
	})

	if len(comm) > nmax {
		comm = comm[:nmax]
	}

	for i := 0; i < len(comm); i++ {
		comm[i] = strings.Trim(comm[i], cutset)

		if k := strings.Index(comm[i], minus); k > 0 {
			comm[i] = comm[i][:k]
		}
	}

	return strings.Join(comm, "-")
}
