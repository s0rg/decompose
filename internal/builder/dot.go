package builder

import (
	"hash/fnv"
	"io"
	"strings"

	"github.com/emicklei/dot"

	"github.com/s0rg/decompose/internal/node"
)

const (
	outPort = "out"
	dotLF   = "&#92;n"
)

// dark28 color scheme from https://www.graphviz.org/doc/info/colors.html
var colors = []string{
	"#1b9e77",
	"#d95f02",
	"#7570b3",
	"#e7298a",
	"#66a61e",
	"#e6ab02",
	"#a6761d",
	"#666666",
}

type DOT struct {
	g        *dot.Graph
	clusters map[string]*dot.Graph
}

func NewDOT() *DOT {
	g := dot.NewGraph(dot.Directed)

	g.Attr("splines", "spline")
	g.Attr("concentrate", dot.Literal("true"))

	return &DOT{
		g:        g,
		clusters: make(map[string]*dot.Graph),
	}
}

func (d *DOT) Name() string {
	return "graphviz-dot"
}

func (d *DOT) AddNode(n *node.Node) error {
	var label, color string

	if n.IsExternal() {
		color = "red"
		label = "external: " + n.Name
	} else {
		color = "black"
		label = n.Name + dotLF + "image: " + n.Image + dotLF + "net: " + strings.Join(n.Networks, ", ")
	}

	if n.Meta != nil {
		if lines, ok := n.FormatMeta(); ok {
			label += dotLF + "info:" + dotLF + strings.Join(lines, dotLF)
		}

		if len(n.Meta.Tags) > 0 {
			label += dotLF + "tags: " + strings.Join(n.Meta.Tags, ",")
		}
	}

	g := d.g

	if n.Cluster != "" {
		sg, ok := d.clusters[n.Cluster]
		if !ok {
			sg = g.Subgraph(n.Cluster, dot.ClusterOption{})
			d.clusters[n.Cluster] = sg
			sg.Node(n.Cluster + "_" + outPort).Label(outPort)
		}

		g = sg
	}

	rb := g.Node(n.ID).Attr(
		"color", color,
	).NewRecordBuilder()

	rb.FieldWithId(label, outPort)
	/*
		    rb.Nesting(func() {
				for i := 0; i < len(n.Ports); i++ {
					p := n.Ports[i]

					rb.FieldWithId(p.Label(), p.ID())
				}
			})
	*/

	_ = rb.Build()

	return nil
}

func (d *DOT) getSrc(id string) (rv dot.Node, out string, ok bool) {
	if rv, ok = d.g.FindNodeById(id); ok {
		return rv, outPort, ok
	}

	sg, ok := d.clusters[id]
	if !ok {
		return
	}

	out = id + "_" + outPort

	rv, ok = sg.FindNodeById(out)

	return rv, out, ok
}

func (d *DOT) getDst(id string, port *node.Port) (rv dot.Node, out string, ok bool) {
	if rv, ok = d.g.FindNodeById(id); ok {
		return rv, port.ID(), ok
	}

	sg, ok := d.clusters[id]
	if !ok {
		return
	}

	out = id + "_" + port.ID()

	if rv, ok = sg.FindNodeById(out); ok {
		return rv, out, ok
	}

	return sg.Node(out).Label(port.Label()), out, true
}

func (d *DOT) AddEdge(e *node.Edge) {
	if e.SrcID == "" || e.DstID == "" { // fast exit, dot doesnt have default cluster
		return
	}

	src, srcPort, ok := d.getSrc(e.SrcID)
	if !ok {
		return
	}

	dst, dstPort, ok := d.getDst(e.DstID, e.Port)
	if !ok {
		return
	}

	label := e.Port.Label()
	color := labelColor(label)

	var edge dot.Edge

	if srcPort != outPort {
		edge = d.g.Edge(src, dst, label)
	} else {
		edge = d.g.EdgeWithPorts(src, dst, srcPort, dstPort, label)
	}

	edge.Attr("color", color).Attr("fontcolor", color)
}

func (d *DOT) Write(w io.Writer) error {
	d.g.Write(w)

	return nil
}

func labelColor(label string) (rv string) {
	h := fnv.New64a()

	_, _ = io.WriteString(h, label)

	hash := int(h.Sum64())

	if hash < 0 {
		hash = -hash
	}

	return colors[hash%len(colors)]
}
