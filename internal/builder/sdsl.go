//go:build !test

package builder

import (
	"errors"
	"fmt"
	"io"

	"github.com/s0rg/decompose/internal/node"
	sdsl "github.com/s0rg/decompose/internal/structurizr"
)

const (
	workspaceName = "de-composed system"
	systemName    = "default"
)

var ErrDuplicate = errors.New("duplicate found")

type Structurizr struct {
	ws *sdsl.Workspace
}

func NewStructurizr() *Structurizr {
	return &Structurizr{
		ws: sdsl.NewWorkspace(workspaceName),
	}
}

func (s *Structurizr) Name() string {
	return "structurizr-dsl"
}

func (s *Structurizr) AddNode(n *node.Node) error {
	system := systemName
	if n.Cluster != "" {
		system = n.Cluster
	}

	state := s.ws.System(system)

	cont, ok := state.AddContainer(n.ID, n.Name)
	if !ok {
		return fmt.Errorf("%w: %s", ErrDuplicate, n.Name)
	}

	cont.Technology = n.Image
	cont.Tags = make([]string, 0, len(n.Ports)+len(n.Networks))

	for _, p := range n.Ports {
		cont.Tags = append(cont.Tags, "listen:"+p.Label())
	}

	for _, n := range n.Networks {
		cont.Tags = append(cont.Tags, "net:"+n)
	}

	if n.IsExternal() {
		cont.Tags = append(cont.Tags, "external")
	}

	if n.Meta != nil {
		cont.Description = n.Meta.Info
		cont.Tags = append(cont.Tags, n.Meta.Tags...)
	}

	return nil
}

func (s *Structurizr) AddEdge(srcID, dstID string, port node.Port) {
	var (
		rel *sdsl.Relation
		ok  bool
	)

	if s.ws.HasSystem(srcID) {
		rel, ok = s.ws.AddRelation(srcID, dstID)
	} else {
		rel, ok = s.ws.System(systemName).AddRelation(srcID, dstID)
	}

	if !ok {
		return
	}

	rel.Tags = append(rel.Tags, port.Label())
}

func (s *Structurizr) Write(w io.Writer) {
	s.ws.Write(w)
}
