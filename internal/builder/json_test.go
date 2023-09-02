package builder_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

type testEnricher struct{}

func (de *testEnricher) Enrich(_ *node.Node) {}

type testAssigner struct{}

func (ta *testAssigner) Assign(_ *node.Node) {}

func TestJSON(t *testing.T) {
	t.Parallel()

	testNode := node.JSON{
		Name:      "test1",
		Networks:  []string{"test"},
		Listen:    []string{"1/udp", "2/tcp"},
		Connected: make(map[string][]string),
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
		Cluster: &testAssigner{},
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.LoadStream(bytes.NewBuffer(jnode)); err != nil {
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
          "listen": ["1/udp"],
          "connected": {"test2":["2/tcp"]}
        }
        {
          "name": "test2",
          "is_external": false,
          "networks": ["test"],
          "listen": ["2/tcp"],
          "connected": {"test1":["1/udp"]}
        }`

	bldr := builder.NewJSON()
	ext := &testEnricher{}

	cfg := &graph.Config{
		Cluster: &testAssigner{},
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.LoadStream(bytes.NewBufferString(raw)); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	var buf bytes.Buffer

	bldr.Write(&buf)

	if strings.Count(buf.String(), "test2") < 2 {
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

func (tb *testBuilder) AddEdge(_, _ string, _ node.Port) {
	tb.Edges++
}

func TestJSONAddBadEdges(t *testing.T) {
	t.Parallel()

	bld := builder.NewJSON()

	_ = bld.AddNode(&node.Node{ID: "1", Name: "1"})
	_ = bld.AddNode(&node.Node{ID: "2", Name: "2"})

	bld.AddEdge("3", "1", node.Port{})
	bld.AddEdge("1", "3", node.Port{})

	var buf bytes.Buffer

	bld.Write(&buf)

	tb := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Cluster: &testAssigner{},
		Builder: tb,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.LoadStream(&buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if tb.Nodes != 2 || tb.Edges != 0 {
		t.Fail()
	}
}
