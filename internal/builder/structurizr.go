//go:build !test

package builder

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

type container struct {
	ID   string
	Name string
}

type Structurizr struct {
	nodes map[string]*container
	edges map[string]map[string][]string
}

func NewStructurizr() *Structurizr {
	return &Structurizr{
		nodes: make(map[string]*container),
		edges: make(map[string]map[string][]string),
	}
}

func (s *Structurizr) AddNode(n *node.Node) error {
	com := &container{
		ID:   strings.ReplaceAll(n.Name, "-", "_"),
		Name: n.Name,
	}

	s.nodes[n.ID] = com

	return nil
}

func (s *Structurizr) AddEdge(srcID, dstID string, port node.Port) {
	csrc, ok := s.nodes[srcID]
	if !ok {
		return
	}

	cdst, ok := s.nodes[dstID]
	if !ok {
		return
	}

	edges, ok := s.edges[csrc.ID]
	if !ok {
		edges = make(map[string][]string)
	}

	dedges, ok := edges[cdst.ID]
	if !ok {
		dedges = make([]string, 0, 1)
	}

	dedges = append(dedges, port.Label())

	edges[cdst.ID] = dedges
	s.edges[csrc.ID] = edges
}

func (s *Structurizr) Write(w io.Writer) {
	bw := bufio.NewWriter(w)

	fmt.Fprintln(bw, `workspace {
    model {
        s = softwareSystem "Software System" {`)

	for _, com := range s.nodes {
		fmt.Fprintf(bw, `%s = container "%s"
`, com.ID, com.Name)
	}

	fmt.Fprintln(bw, "}")

	for srcID, edges := range s.edges {
		for dstID, ports := range edges {
			for _, label := range ports {
				fmt.Fprintf(bw, `%s -> %s "%s"
`, srcID, dstID, label)
			}
		}
	}

	fmt.Fprintln(bw, `}

views {
        container s {
            include *
            autoLayout lr
        }
    }
}`)

	_ = bw.Flush()
}
