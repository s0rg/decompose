package builder_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/node"
)

func TestYAMLGolden(t *testing.T) {
	t.Parallel()

	bld := builder.NewYAML()

	_ = bld.AddNode(&node.Node{
		ID:    "node-1",
		Name:  "1",
		Image: "node-image",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
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
		Volumes: []*node.Volume{
			{Type: "volume", Src: "src", Dst: "dst"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:    "node-2",
		Name:  "2",
		Image: "node-image",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
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
		Volumes: []*node.Volume{
			{Type: "volume", Src: "src2", Dst: "dst2"},
		},
	})

	_ = bld.AddNode(&node.Node{
		ID:   "2",
		Name: "2",
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
		SrcID: "node-2",
		DstID: "node-1",
		Port:  &node.Port{Kind: "tcp", Value: 2},
	})

	bld.AddEdge(&node.Edge{
		SrcID: "node-1",
		DstID: "node-2",
		Port:  &node.Port{Kind: "tcp", Value: 2},
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

	var buf bytes.Buffer

	bld.Write(&buf)

	got := buf.String()
	want := golden(t, bld.Name(), got)

	if got != want {
		t.Errorf("Want:\n%s\nGot:\n%s", want, got)
	}
}

func TestYAMLWriteError(t *testing.T) {
	t.Parallel()

	bldr := builder.NewYAML()
	testErr := errors.New("test-error")
	errW := &errWriter{Err: testErr}

	_ = bldr.AddNode(&node.Node{
		ID:       "#",
		Name:     "#",
		Image:    "node-image",
		Ports:    &node.Ports{},
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 1",
			Tags: []string{"1"},
		},
		Container: node.Container{
			Cmd: []string{"echo", "'test 1'"},
			Env: []string{"FOO=1"},
		},
		Volumes: []*node.Volume{
			{Type: "volume", Src: "src", Dst: "dst"},
		},
	})

	if err := bldr.Write(errW); !errors.Is(err, testErr) {
		t.Log(err)
	}
}
