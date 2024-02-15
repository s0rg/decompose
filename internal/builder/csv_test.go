package builder_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/node"
)

func TestCSVGolden(t *testing.T) {
	t.Parallel()

	bld := builder.NewCSV()

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
		Container: node.Container{
			Cmd: []string{"echo", "'test 1'"},
			Env: []string{"FOO=1"},
		},
	})
	_ = bld.AddNode(&node.Node{
		ID:   "node-2",
		Name: "2",
		Ports: node.Ports{
			{Kind: "tcp", Value: 2},
		},
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
		Ports: node.Ports{
			{Kind: "tcp", Value: 3},
		},
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

	bld.AddEdge("node-2", "node-1", &node.Port{Kind: "tcp", Value: 1})
	bld.AddEdge("node-2", "node-1", &node.Port{Kind: "tcp", Value: 2})
	bld.AddEdge("node-2", "node-1", &node.Port{Kind: "tcp", Value: 3})

	bld.AddEdge("node-2", "node-3", &node.Port{Kind: "tcp", Value: 3})

	bld.AddEdge("node-1", "node-3", &node.Port{Kind: "tcp", Value: 3})
	bld.AddEdge("node-1", "node-2", &node.Port{Kind: "tcp", Value: 2})

	bld.AddEdge("node-1", "3", &node.Port{Kind: "tcp", Value: 3})
	bld.AddEdge("3", "node-1", &node.Port{Kind: "tcp", Value: 3})

	var buf bytes.Buffer

	bld.Write(&buf)

	got := buf.String()
	want := golden(t, bld.Name(), got)

	if got != want {
		t.Errorf("Want:\n%s\nGot:\n%s", want, got)
	}
}

func TestCSVWriteError(t *testing.T) {
	t.Parallel()

	bldr := builder.NewCSV()
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
