package builder

import (
	"errors"
	"fmt"
	"io"
	"strings"

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
		ws: sdsl.NewWorkspace(workspaceName, systemName),
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

	state.Tags = append(state.Tags, n.Name)

	cont, ok := state.AddContainer(n.ID, n.Name)
	if !ok {
		return fmt.Errorf("%w: %s", ErrDuplicate, n.Name)
	}

	cont.Technology = n.Image
	/*
		    cont.Tags = make([]string, 0, len(n.Ports)+len(n.Networks))

			for _, p := range n.Ports {
				cont.Tags = append(cont.Tags, "listen:"+p.Label())
			}
	*/

	for _, n := range n.Networks {
		cont.Tags = append(cont.Tags, "net:"+n)
	}

	if n.IsExternal() {
		cont.Tags = append(cont.Tags, "external")
	}

	if lines, ok := n.FormatMeta(); ok {
		cont.Description = strings.Join(lines, " \\\n")

		cont.Tags = append(cont.Tags, n.Meta.Tags...)
	}

	return nil
}

func (s *Structurizr) AddEdge(e *node.Edge) {
	var (
		rel *sdsl.Relation
		ok  bool
	)

	switch {
	case e.SrcID == "":
		e.SrcID = systemName
	case e.DstID == "":
		e.DstID = systemName
	}

	if s.ws.HasSystem(e.SrcID) {
		rel, ok = s.ws.AddRelation(e.SrcID, e.DstID)
	} else {
		rel, ok = s.ws.System(systemName).AddRelation(e.SrcID, e.DstID)
	}

	if !ok {
		return
	}

	rel.Tags = append(rel.Tags, e.Port.Label())
}

func (s *Structurizr) Write(w io.Writer) error {
	s.ws.Write(w)

	return nil
}
