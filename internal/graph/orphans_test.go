package graph_test

import (
	"io"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

func TestOrphans(t *testing.T) {
	t.Parallel()

	tb := testNamedBuilder{}
	op := graph.NewOrphansInspector(&tb)

	name := op.Name()

	if !strings.Contains(name, tb.Name()) {
		t.Fail()
	}

	op.AddNode(&node.Node{ID: "1"})
	op.AddNode(&node.Node{ID: "2"})
	op.AddNode(&node.Node{ID: "3"})

	op.AddEdge(&node.Edge{SrcID: "1", DstID: "3"})
	op.AddEdge(&node.Edge{SrcID: "3", DstID: "1"})

	_ = op.Write(io.Discard)

	if tb.Edges != 2 || tb.Nodes != 2 {
		t.Fail()
	}
}
