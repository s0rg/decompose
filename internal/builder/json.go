package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/s0rg/decompose/internal/node"
)

type JSON struct {
	state map[string]*node.JSON
}

func NewJSON() *JSON {
	return &JSON{
		state: make(map[string]*node.JSON),
	}
}

func (j *JSON) Name() string {
	return "json-stream"
}

func (j *JSON) AddNode(n *node.Node) error {
	j.state[n.ID] = n.ToJSON()

	return nil
}

func (j *JSON) AddEdge(srcID, dstID string, port *node.Port) {
	src, ok := j.state[srcID]
	if !ok {
		return
	}

	dst, ok := j.state[dstID]
	if !ok {
		return
	}

	con, ok := src.Connected[dst.Name]
	if !ok {
		con = make([]string, 0, 1)
	}

	src.Connected[dst.Name] = append(con, port.Label())
}

func (j *JSON) Sorted(fn func(*node.JSON, bool)) {
	nodes := make([]*node.JSON, 0, len(j.state))

	for _, n := range j.state {
		nodes = append(nodes, n)
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	for i, n := range nodes {
		fn(n, i == len(nodes)-1)
	}
}

func (j *JSON) Write(w io.Writer) (err error) {
	jw := json.NewEncoder(w)
	jw.SetIndent("", "  ")

	j.Sorted(func(n *node.JSON, _ bool) {
		err = jw.Encode(n)
	})

	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	return nil
}
