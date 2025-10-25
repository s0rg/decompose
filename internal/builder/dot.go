package builder

import (
	"io"
	"slices"

	"github.com/emicklei/dot"

	"github.com/s0rg/decompose/internal/node"
)

type DOT struct {
	g     *dot.Graph
	edges map[string]map[string][]string
}

func NewDOT() *DOT {
	g := dot.NewGraph(dot.Directed)

	return &DOT{
		g:     g,
		edges: make(map[string]map[string][]string),
	}
}

func (d *DOT) Name() string {
	return "graphviz-dot"
}

func (d *DOT) AddNode(n *node.Node) error {
	label, color := renderNode(n)

	d.g.Node(n.ID).Attr(
		"color", color,
	).Label(label)

	return nil
}

func (d *DOT) AddEdge(e *node.Edge) {
	if _, ok := d.g.FindNodeById(e.SrcID); !ok {
		return
	}

	if _, ok := d.g.FindNodeById(e.DstID); !ok {
		return
	}

	d.addEdge(e.SrcID, e.DstID, e.Port.Label())
}

func (d *DOT) Write(w io.Writer) error {
	d.buildEdges()
	d.g.Write(w)

	return nil
}

func (d *DOT) addEdge(src, dst, label string) {
	dmap, ok := d.edges[src]
	if !ok {
		dmap = make(map[string][]string)
		d.edges[src] = dmap
	}

	dmap[dst] = append(dmap[dst], label)
}

func (d *DOT) buildEdges() {
	order := make([]string, 0, len(d.edges))

	for k := range d.edges {
		order = append(order, k)
	}

	slices.Sort(order)

	for _, srcID := range order {
		src, _ := d.g.FindNodeById(srcID)

		dmap := d.edges[srcID]
		dorder := make([]string, 0, len(dmap))

		for k := range dmap {
			dorder = append(dorder, k)
		}

		slices.Sort(dorder)

		for _, dstID := range dorder {
			dst, _ := d.g.FindNodeById(dstID)
			ports := dmap[dstID]

			if tmp, ok := d.edges[dstID]; ok {
				if dports, ok := tmp[srcID]; ok {
					ports = append(ports, dports...)

					delete(tmp, srcID)
				}
			}

			d.g.Edge(src, dst, ports...)
		}
	}
}

func renderNode(n *node.Node) (label, color string) {
	label, color = n.Name, "black"

	if n.IsExternal() {
		color = "gray"
		label = "external: " + n.Name
	}

	return label, color
}
