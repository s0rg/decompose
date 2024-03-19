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
	edges    map[string]map[string][]string
}

func NewDOT() *DOT {
	g := dot.NewGraph(dot.Directed)

	g.Attr("splines", "spline")
	g.Attr("concentrate", dot.Literal("true"))

	return &DOT{
		g:        g,
		clusters: make(map[string]*dot.Graph),
		edges:    make(map[string]map[string][]string),
	}
}

func (d *DOT) Name() string {
	return "graphviz-dot"
}

func (d *DOT) AddNode(n *node.Node) error {
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

	label, color := renderNode(n)

	rb := g.Node(n.ID).Attr(
		"color", color,
	).NewRecordBuilder()

	rb.FieldWithId(label, outPort)

	if n.Ports.Len() > 0 {
		rb.Nesting(func() {
			n.Ports.Iter(func(process string, _ []*node.Port) {
				rb.FieldWithId(process, portID(n.ID, process))
			})
		})
	}

	// this cannot return error, thus error case cannot be tested :(
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

func (d *DOT) getDst(edge *node.Edge) (rv dot.Node, out string, ok bool) {
	dstID := portID(edge.DstID, edge.DstName)

	if rv, ok = d.g.FindNodeById(dstID); ok {
		return rv, dstID, ok
	}

	if rv, ok = d.g.FindNodeById(edge.DstID); ok {
		return rv, edge.DstID, ok
	}

	sg, ok := d.clusters[edge.DstID]
	if !ok {
		return
	}

	if rv, ok = sg.FindNodeById(dstID); ok {
		return rv, dstID, ok
	}

	return sg.Node(out).Label(edge.Port.Label()), out, true
}

func (d *DOT) AddEdge(e *node.Edge) {
	if e.SrcID == "" || e.DstID == "" { // fast exit, dot doesnt have default cluster
		return
	}

	src, srcPort, ok := d.getSrc(e.SrcID)
	if !ok {
		return
	}

	dst, dstPort, ok := d.getDst(e)
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

func portID(id, name string) (rv string) {
	return "port_" + id + "_" + name
}

func renderNode(n *node.Node) (label, color string) {
	var sb strings.Builder

	if n.IsExternal() {
		color = "gray"

		sb.WriteString("external: ")
	} else {
		color = "black"
	}

	sb.WriteString(n.Name)
	sb.WriteString(dotLF)

	if n.Image != "" {
		sb.WriteString("image: ")
		sb.WriteString(n.Image)
		sb.WriteString(dotLF)
	}

	if len(n.Networks) > 0 {
		sb.WriteString("nets: ")
		sb.WriteString(strings.Join(n.Networks, ", "))
		sb.WriteString(dotLF)
	}

	if n.Meta != nil {
		if lines, ok := n.FormatMeta(); ok {
			sb.WriteString("meta: ")
			sb.WriteString(dotLF)
			sb.WriteString(strings.Join(lines, dotLF))
			sb.WriteString(dotLF)
		}

		if len(n.Meta.Tags) > 0 {
			sb.WriteString("tags: ")
			sb.WriteString(strings.Join(n.Meta.Tags, ","))
			sb.WriteString(dotLF)
		}
	}

	return sb.String(), color
}
