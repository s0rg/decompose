//go:build !test

package builder

import (
	"fmt"
	"hash/fnv"
	"io"
	"strings"

	"github.com/emicklei/dot"

	"github.com/s0rg/decompose/internal/node"
)

const outPort = "out"

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
		label = fmt.Sprintf("external: %s", n.Name)
	} else {
		color = "black"
		label = fmt.Sprintf(
			"%s&#92;nimage: %s&#92;nnet: %s",
			n.Name,
			n.Image,
			strings.Join(n.Networks, ", "),
		)
	}

	if n.Meta != nil {
		if len(n.Meta.Info) > 0 {
			label += "&#92;ninfo: " + n.Meta.Info
		}

		if len(n.Meta.Tags) > 0 {
			label += "&#92;ntags: " + strings.Join(n.Meta.Tags, ",")
		}
	}

	g := d.g

	if n.Cluster != "" {
		sg, ok := d.clusters[n.Cluster]
		if !ok {
			sg = g.Subgraph(n.Cluster, dot.ClusterOption{})
			d.clusters[n.Cluster] = sg
		}

		g = sg
	}

	rb := g.Node(n.ID).Attr("color", color).NewRecordBuilder()

	rb.FieldWithId(label, outPort)
	rb.Nesting(func() {
		for i := 0; i < len(n.Ports); i++ {
			p := &n.Ports[i]

			rb.FieldWithId(p.Label(), p.ID())
		}
	})

	if err := rb.Build(); err != nil {
		return fmt.Errorf("node for %s: %w", n.ID, err)
	}

	return nil
}

func (d *DOT) AddEdge(srcID, dstID string, port node.Port) {
	src, ok := d.g.FindNodeById(srcID)
	if !ok {
		return
	}

	dst, ok := d.g.FindNodeById(dstID)
	if !ok {
		return
	}

	color := labelColor(
		min(srcID, dstID) + max(srcID, dstID) + port.Label(),
	)

	d.g.
		EdgeWithPorts(src, dst, outPort, port.ID(), port.Label()).
		Attr("color", color).
		Attr("fontcolor", color)
}

func (d *DOT) Write(w io.Writer) {
	d.g.Write(w)
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
