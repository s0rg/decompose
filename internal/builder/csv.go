package builder

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

var csvHeader = []string{
	"service", "listen", "outbounds",
}

type CSV struct {
	j *JSON
}

func NewCSV() *CSV {
	return &CSV{
		j: NewJSON(),
	}
}

func (c *CSV) Name() string {
	return "csv"
}

func (c *CSV) AddNode(n *node.Node) error {
	return c.j.AddNode(n)
}

func (c *CSV) AddEdge(srcID, dstID string, port *node.Port) {
	c.j.AddEdge(srcID, dstID, port)
}

func (c *CSV) Write(w io.Writer) error {
	cw := csv.NewWriter(w)
	cw.UseCRLF = true

	_ = cw.Write(csvHeader)

	c.j.Sorted(func(n *node.JSON, _ bool) {
		_ = cw.Write([]string{
			n.Name,
			strings.Join(n.Listen, "\r\n"),
			renderOutbounds(n.Connected),
		})
	})

	cw.Flush()

	if err := cw.Error(); err != nil {
		return fmt.Errorf("fail: %w", err)
	}

	return nil
}

func renderOutbounds(conns map[string][]string) (rv string) {
	var b strings.Builder

	for k, v := range conns {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(strings.Join(v, "; "))
		b.WriteString("\r\n")
	}

	return b.String()
}
