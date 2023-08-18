//go:build !test

package builder

import (
	"fmt"
	"hash/fnv"
	"io"
	"strconv"
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
	g *dot.Graph
}

func NewDOT() *DOT {
	g := dot.NewGraph(dot.Directed)

	g.Attr("splines", "spline")
	g.Attr("concentrate", dot.Literal("true"))

	return &DOT{g: g}
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

	rb := d.g.Node(n.ID).Attr("color", color).NewRecordBuilder()
	rb.FieldWithId(label, outPort)
	rb.Nesting(func() {
		for i := 0; i < len(n.Ports); i++ {
			p := &n.Ports[i]

			rb.FieldWithId(p.Label(), strconv.Itoa(p.Value))
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
		EdgeWithPorts(src, dst, outPort, port.String(), port.Label()).
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
