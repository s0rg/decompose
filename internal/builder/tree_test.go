package builder_test

import (
	"bytes"
	"testing"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/node"
)

func TestTreeGolden(t *testing.T) {
	t.Parallel()

	bld := builder.NewTree()

	_ = bld.AddNode(&node.Node{
		ID:    "node-1",
		Name:  "1",
		Image: "node-image",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1"},
			{Kind: "tcp", Value: "2"},
		}...),
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 1",
			Tags: []string{"1"},
		},
		Container: node.Container{
			Cmd: []string{"echo", "'test 1'"},
			Env: []string{"FOO=1"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:   "node-2",
		Name: "2",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "2"},
		}...),
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 2",
			Tags: []string{"2"},
		},
		Container: node.Container{
			Cmd: []string{"echo", "'test 2'"},
			Env: []string{"FOO=2"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:   "node-3",
		Name: "3",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "3"},
		}...),
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 3",
			Tags: []string{"3"},
		},
		Container: node.Container{
			Cmd: []string{"echo", "'test 3'"},
			Env: []string{"FOO=3"},
		},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: "1"},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: "2"},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: "3"},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-2",
		DstID: "node-3",
		Port:  &node.Port{Kind: "tcp", Value: "3"},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "node-3",
		Port:  &node.Port{Kind: "tcp", Value: "3"},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "node-2",
		Port:  &node.Port{Kind: "tcp", Value: "2"},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "3",
		Port:  &node.Port{Kind: "tcp", Value: "3"},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "3",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: "3"},
	})

	var buf bytes.Buffer

	bld.Write(&buf)

	got := buf.String()
	want := golden(t, bld.Name(), got)

	if got != want {
		t.Errorf("Want:\n%s\nGot:\n%s", want, got)
	}
}
