package builder

import (
	"encoding/json"
	"io"

	"github.com/s0rg/decompose/internal/node"
)

type nodeJSON struct {
	Name        string              `json:"name"`
	Image       *string             `json:"image,omitempty"`
	IsExternal  bool                `json:"is_external"`
	Ports       []string            `json:"ports"`
	Connections map[string][]string `json:"connected"`
}

type JSON struct {
	state map[string]*nodeJSON
}

func NewJSON() *JSON {
	return &JSON{
		state: make(map[string]*nodeJSON),
	}
}

func (j *JSON) AddNode(n *node.Node) error {
	jn := &nodeJSON{
		Name:        n.Name,
		IsExternal:  n.IsExternal(),
		Ports:       make([]string, len(n.Ports)),
		Connections: make(map[string][]string),
	}

	if n.Image != "" {
		jn.Image = &n.Image
	}

	for i := 0; i < len(n.Ports); i++ {
		jn.Ports[i] = n.Ports[i].Label()
	}

	j.state[n.ID] = jn

	return nil
}

func (j *JSON) AddEdge(srcID, dstID string, port node.Port) {
	src, ok := j.state[srcID]
	if !ok {
		return
	}

	dst, ok := j.state[dstID]
	if !ok {
		return
	}

	con, ok := src.Connections[dst.Name]
	if !ok {
		con = make([]string, 0, 1)
	}

	src.Connections[dst.Name] = append(con, port.Label())
}

func (j *JSON) Write(w io.Writer) {
	jw := json.NewEncoder(w)
	jw.SetIndent("", "  ")

	for _, cnt := range j.state {
		_ = jw.Encode(cnt)
	}
}
