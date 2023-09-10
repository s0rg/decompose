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
		Ports: node.Ports{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		},
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 1",
			Tags: []string{"1"},
		},
		Process: &node.Process{
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
		Ports: node.Ports{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		},
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 2",
			Tags: []string{"2"},
		},
		Process: &node.Process{
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
		Ports: node.Ports{
			{Kind: "tcp", Value: 2},
		},
	})

	_ = bld.Name()

	bld.AddEdge("2", "node-1", &node.Port{Kind: "tcp", Value: 1})
	bld.AddEdge("2", "node-1", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("2", "node-1", &node.Port{Kind: "tcp", Value: 3})

	bld.AddEdge("node-2", "node-1", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("node-1", "node-2", &node.Port{Kind: "tcp", Value: 2})

	bld.AddEdge("node-1", "2", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("node-1", "2", &node.Port{Kind: "tcp", Value: 3})

	bld.AddEdge("node-1", "3", &node.Port{Kind: "tcp", Value: 3})
	bld.AddEdge("3", "node-1", &node.Port{Kind: "tcp", Value: 3})

	var buf bytes.Buffer

	bld.Write(&buf)

	got := buf.String()
	want := golden(t, "yaml", got)

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
		ID:    "#",
		Name:  "#",
		Image: "node-image",
		Ports: node.Ports{
			{Kind: "tcp", Value: 1},
			{Kind: "tcp", Value: 2},
		},
		Networks: []string{"test-net"},
		Meta: &node.Meta{
			Info: "info 1",
			Tags: []string{"1"},
		},
		Process: &node.Process{
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
