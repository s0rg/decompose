package builder_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

type testEnricher struct{}

func (de *testEnricher) Enrich(_ *node.Node) {}

type errWriter struct {
	Err error
}

func (ew *errWriter) Write(_ []byte) (n int, err error) { return 0, ew.Err }

func TestJSON(t *testing.T) {
	t.Parallel()

	testNode := node.JSON{
		Name:     "test1",
		Networks: []string{"test"},
		Listen: map[string][]*node.Port{
			"foo": {
				&node.Port{Kind: "tcp", Value: 2},
				&node.Port{Kind: "udp", Value: 1},
			},
		},
		Tags:      []string{},
		Connected: make(map[string][]*node.Connection),
		Volumes:   []*node.Volume{},
	}

	jnode, err := json.Marshal(testNode)
	if err != nil {
		t.Fatal("marshal err=", err)
	}

	var rawCompact bytes.Buffer

	if err := json.Compact(&rawCompact, jnode); err != nil {
		t.Fatal("raw compact err=", err)
	}

	bldr := builder.NewJSON()

	if bldr.Name() != "json-stream" {
		t.Fail()
	}

	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(bytes.NewBuffer(jnode)); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	var buf bytes.Buffer

	bldr.Write(&buf)

	var resCompact bytes.Buffer

	if err := json.Compact(&resCompact, buf.Bytes()); err != nil {
		t.Fatal("res compact err=", err)
	}

	if rawCompact.String() != resCompact.String() {
		t.Log("want:", rawCompact.String())
		t.Log("got:", resCompact.String())

		t.Fail()
	}
}

func TestJSONAddEdge(t *testing.T) {
	t.Parallel()

	const raw = `{
          "name": "test1",
          "is_external": false,
          "networks": ["test"],
          "listen": {"foo":[
            {"kind": "udp", "value": 1}
           ]},
          "connected": {"test2":[{"src": "foo", "dst": "bar", "port": {"kind": "tcp", "value": 2}}]}
        }
        {
          "name": "test2",
          "is_external": false,
          "networks": ["test"],
          "listen": {"bar": [
            {"kind": "tcp", "value": 2}
          ]},
          "connected": {"test1":[{"src": "bar", "dst": "foo", "port": {"kind": "udp", "value": 1}}]}
        }`

	bldr := builder.NewJSON()
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(bytes.NewBufferString(raw)); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	var buf bytes.Buffer

	bldr.Write(&buf)

	if strings.Count(buf.String(), "test2") < 2 {
		t.Log(buf)
		t.Fail()
	}
}

type testBuilder struct {
	Nodes int
	Edges int
}

func (tb *testBuilder) AddNode(_ *node.Node) error {
	tb.Nodes++

	return nil
}

func (tb *testBuilder) AddEdge(_ *node.Edge) {
	tb.Edges++
}

func TestJSONAddBadEdges(t *testing.T) {
	t.Parallel()

	bld := builder.NewJSON()

	_ = bld.AddNode(&node.Node{ID: "1", Name: "1", Ports: &node.Ports{}})
	_ = bld.AddNode(&node.Node{ID: "2", Name: "2", Ports: &node.Ports{}})

	bld.AddEdge(&node.Edge{SrcID: "3", DstID: "1", Port: &node.Port{}})
	bld.AddEdge(&node.Edge{SrcID: "1", DstID: "3", Port: &node.Port{}})

	var buf bytes.Buffer

	bld.Write(&buf)

	tb := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: tb,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(&buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if tb.Nodes != 2 || tb.Edges != 0 {
		t.Fail()
	}
}

func TestJSONWriteError(t *testing.T) {
	t.Parallel()

	bldr := builder.NewJSON()
	testErr := errors.New("test-error")
	errW := &errWriter{Err: testErr}

	_ = bldr.AddNode(&node.Node{ID: "1", Name: "1", Ports: &node.Ports{}})

	if err := bldr.Write(errW); !errors.Is(err, testErr) {
		t.Fail()
	}
}
