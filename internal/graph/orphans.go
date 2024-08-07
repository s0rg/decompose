package graph

import (
	"fmt"
	"io"

	"github.com/s0rg/decompose/internal/node"

	"github.com/s0rg/set"
)

const orphansName = "no-orphans"

type OrphansInspector struct {
	b NamedBuilderWriter
	o set.Unordered[string]
	n []*node.Node
	e []*node.Edge
}

func NewOrphansInspector(b NamedBuilderWriter) *OrphansInspector {
	return &OrphansInspector{
		b: b,
		o: make(set.Unordered[string]),
	}
}

func (o *OrphansInspector) Name() string {
	return o.b.Name() + " " + orphansName
}

func (o *OrphansInspector) AddNode(n *node.Node) error {
	o.n = append(o.n, n)
	o.o.Add(n.ID)

	return nil
}

func (o *OrphansInspector) AddEdge(e *node.Edge) {
	o.e = append(o.e, e)
	o.o.Del(e.SrcID)
	o.o.Del(e.DstID)
}

func (o *OrphansInspector) Write(w io.Writer) (err error) {
	for _, n := range o.n {
		if o.o.Has(n.ID) {
			continue
		}

		if err = o.b.AddNode(n); err != nil {
			return fmt.Errorf("no-orphans add node: %w", err)
		}
	}

	for _, e := range o.e {
		o.b.AddEdge(e)
	}

	if err = o.b.Write(w); err != nil {
		return fmt.Errorf("no-orphans write: %w", err)
	}

	return nil
}
