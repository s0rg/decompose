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
		Ports: node.Ports{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		},
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 1",
			Tags: []string{"1"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:      "node-2",
		Name:    "2",
		Image:   "node-image",
		Cluster: "c1",
		Ports: node.Ports{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		},
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
		Ports: node.Ports{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		},
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
		Ports: node.Ports{
			{Kind: "tcp", Value: 2},
		},
	})

	_ = bld.Name()

	bld.AddEdge("2", "node-1", &node.Port{Kind: "tcp", Value: 1})
	bld.AddEdge("2", "node-1", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("2", "node-1", &node.Port{Kind: "tcp", Value: 3})

	bld.AddEdge("node-1", "2", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("node-1", "2", &node.Port{Kind: "tcp", Value: 3})

	bld.AddEdge("node-1", "3", &node.Port{Kind: "tcp", Value: 3})
	bld.AddEdge("3", "node-1", &node.Port{Kind: "tcp", Value: 3})

	bld.AddEdge("c1", "c2", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("c1", "c2", &node.Port{Kind: "tcp", Value: 2})

	bld.AddEdge("c2", "c1", &node.Port{Kind: "tcp", Value: 1})
	bld.AddEdge("c2", "c1", &node.Port{Kind: "tcp", Value: 1})

	bld.AddEdge("c1", "", &node.Port{})
	bld.AddEdge("", "c2", &node.Port{})

	bld.AddEdge("node-1", "node-2", &node.Port{Kind: "tcp", Value: 1})
	bld.AddEdge("node-1", "node-2", &node.Port{Kind: "tcp", Value: 1})

	bld.AddEdge("node-2", "node-1", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("node-2", "node-1", &node.Port{Kind: "tcp", Value: 2})

	var buf bytes.Buffer

	bld.Write(&buf)

	got := buf.String()
	want := golden(t, "dot", got)

	if got != want {
		t.Errorf("Want:\n%s\nGot:\n%s", want, got)
	}
}
