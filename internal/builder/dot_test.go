package builder_test

import (
	"bytes"
	"testing"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/node"
)

func TestDOTGolden(t *testing.T) {
	t.Parallel()

	bld := builder.NewDOT()

	_ = bld.AddNode(&node.Node{
		ID:      "node-1",
		Name:    "1",
		Image:   "node-image",
		Cluster: "c1",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		}...),
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 1",
			Docs: "docs-url",
			Repo: "repo-url",
			Tags: []string{"1"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:      "node-2",
		Name:    "2",
		Image:   "node-image",
		Cluster: "c1",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		}...),
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 2",
			Tags: []string{"2"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:      "node-3",
		Name:    "3",
		Image:   "node-image",
		Cluster: "c3",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		}...),
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 3",
			Tags: []string{"3"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:      "2",
		Name:    "2",
		Cluster: "c2",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: 2},
		}...),
	})

	bld.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: 1},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: 2},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: 3},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "2",
		Port:  &node.Port{Kind: "tcp", Value: 2},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "2",
		Port:  &node.Port{Kind: "tcp", Value: 3},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "3",
		Port:  &node.Port{Kind: "tcp", Value: 3},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "3",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: 3},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "c1",
		DstID: "c2",
		Port:  &node.Port{Kind: "tcp", Value: 2},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "c1",
		DstID: "c2",
		Port:  &node.Port{Kind: "tcp", Value: 2},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "c2",
		DstID: "c1",
		Port:  &node.Port{Kind: "tcp", Value: 1},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "c2",
		DstID: "c1",
		Port:  &node.Port{Kind: "tcp", Value: 1},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "c1",
		DstID: "",
		Port:  &node.Port{},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "",
		DstID: "c2",
		Port:  &node.Port{},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "node-2",
		Port:  &node.Port{Kind: "tcp", Value: 1},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "node-2",
		Port:  &node.Port{Kind: "tcp", Value: 1},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: 2},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: 2},
	})

	var buf bytes.Buffer

	bld.Write(&buf)

	got := buf.String()
	want := golden(t, bld.Name(), got)

	if got != want {
		t.Errorf("Want:\n%s\nGot:\n%s", want, got)
	}
}
