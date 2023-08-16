package builder

import (
	"fmt"
	"io"
	"strconv"

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
	return &DOT{
		g: dot.NewGraph(dot.Directed),
	}
}

func (d *DOT) AddNode(n *node.Node) error {
	color := "black"
	if n.IsExternal() {
		color = "red"
	}

	rb := d.g.Node(n.ID).Attr("color", color).NewRecordBuilder()
	rb.FieldWithId(n.Name+"&#92;n"+n.Image, outPort)
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

	color := portToColor(port.Value)

	d.g.
		EdgeWithPorts(src, dst, outPort, port.String(), port.Label()).
		Attr("color", color).
		Attr("fontcolor", color)
}

func (d *DOT) Write(w io.Writer) {
	d.g.Write(w)
}

func portToColor(i int) (rv string) {
	return colors[i%len(colors)]
}
