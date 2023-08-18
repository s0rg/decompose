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

func TestJSON(t *testing.T) {
	t.Parallel()

	const raw = `{
          "name": "test1",
          "is_external": false,
          "networks": ["test"],
          "listen": ["1/udp"],
          "connected": {}
        }
    `

	var rawCompact bytes.Buffer

	if err := json.Compact(&rawCompact, []byte(raw)); err != nil {
		t.Fatal("raw compact err=", err)
	}

	ldr := graph.NewLoader("", "", false)

	if err := ldr.LoadStream(bytes.NewBufferString(raw)); err != nil {
		t.Fatal("load err=", err)
	}

	bld := builder.NewJSON()

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	var buf bytes.Buffer

	bld.Write(&buf)

	var resCompact bytes.Buffer

	if err := json.Compact(&resCompact, buf.Bytes()); err != nil {
		t.Fatal("res compact err=", err)
	}

	if rawCompact.String() != resCompact.String() {
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

	ldr := graph.NewLoader("", "", false)

	if err := ldr.LoadStream(bytes.NewBufferString(raw)); err != nil {
		t.Fatal("load err=", err)
	}

	bld := builder.NewJSON()

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	var buf bytes.Buffer

	bld.Write(&buf)

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

	ldr := graph.NewLoader("", "", false)

	if err := ldr.LoadStream(&buf); err != nil {
		t.Fatal("load err=", err)
	}

	tb := &testBuilder{}

	if err := ldr.Build(tb); err != nil {
		t.Fatal("build err=", err)
	}

	if tb.Nodes != 2 || tb.Edges != 0 {
		t.Fail()
	}
}
