package cluster

import (
	"fmt"
	"io"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

type Layers struct {
	r set.Unordered[string]
	g connGraph
	b graph.NamedBuilderWriter
	p float64
}

func NewLayers(
	b graph.NamedBuilderWriter,
	p float64,
) *Layers {
	return &Layers{
		r: make(set.Unordered[string]),
		g: make(connGraph),
		b: b,
		p: p,
	}
}

func (l *Layers) Name() string {
	return l.b.Name() + " layers-clustered"
}

func (l *Layers) AddNode(n *node.Node) error {
	if n.IsExternal() {
		l.r.Add(n.ID)

		return nil
	}

	l.g.AddNode(n)

	return nil
}

func (l *Layers) AddEdge(src, dst string, _ *node.Port) {
	if l.r.Has(src) || l.r.Has(dst) {
		return
	}

	l.g.AddEdge(src, dst)
}

func (l *Layers) Write(w io.Writer) error {
	var (
		layer          []string
		gtotal, etotal int
	)

	for i := 0; ; i++ {
		layer = l.g.NextLayer(layer)
		if len(layer) == 0 {
			break
		}

		grp := NewGrouper(l.p)

		for _, name := range layer {
			grp.Add(name, l.g[name].Ports)
		}

		fmt.Println("layer: ", i)
		fmt.Println("  initial:", grp.Groups())

		grp.Compress()

		// fmt.Println(grp)

		fmt.Println("  compressed:", grp.Groups())

		count := grp.Count()

		etotal += count
		gtotal += grp.Groups()

		fmt.Println("  elements:", count)
		fmt.Println("===============================")
	}

	fmt.Println("total:")
	fmt.Println("  elements:", etotal)
	fmt.Println("  groups:", gtotal)

	return nil
}
