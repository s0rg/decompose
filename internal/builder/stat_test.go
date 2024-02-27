package builder_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/cluster"
	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

func TestStat(t *testing.T) {
	t.Parallel()

	bldr := builder.NewStat()

	if bldr.Name() != "graph-stats" {
		t.Fail()
	}

	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	const raw = `{
    "name": "test1",
    "is_external": false,
    "networks": ["test"],
    "listen": {"1": ["1/udp", "1/tcp"]},
    "connected": {
        "test2":[
            {"src": "1", "dst": "2", "port": "2/tcp"},
            {"src": "1", "dst": "2", "port": "2/udp"}
        ],
        "test3":[
            {"src": "1", "dst": "3", "port": "3/tcp"}
        ]
      }
    }
    {
    "name": "test2",
    "is_external": false,
    "networks": ["test"],
    "listen": {"2":["2/tcp", "2/udp"]},
    "connected": {
        "test1":[
            {"src": "2", "dst": "1", "port": "1/tcp"},
            {"src": "2", "dst": "1", "port": "1/udp"}
        ],
        "test3": [
            {"src": "2", "dst": "3", "port": "3/udp"}
        ]
      }
    }
    {
    "name": "test3",
    "is_external": true,
    "networks": ["test"],
    "listen": {"3":["3/tcp", "3/udp"]},
    "connected": {}
    }`

	if err := ldr.FromReader(bytes.NewBufferString(raw)); err != nil {
		t.Fatal("load err=", err)
	}

	bldr.AddEdge(&node.Edge{SrcID: "test1", DstID: "bad-id"})
	bldr.AddEdge(&node.Edge{SrcID: "bad-id", DstID: "test1"})

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	var buf bytes.Buffer

	bldr.Write(&buf)

	res := buf.String()

	if !strings.Contains(res, "Nodes: 2") {
		t.Fail()
	}

	if !strings.Contains(res, "total: 4 uniq: 1") {
		t.Fail()
	}

	if !strings.Contains(res, "Externals: 1") {
		t.Fail()
	}

	if strings.Count(res, "1/tcp") != 1 {
		t.Fail()
	}

	if strings.Count(res, "2/tcp") != 1 {
		t.Fail()
	}

	if strings.Count(res, "1/udp") != 1 {
		t.Fail()
	}

	if strings.Count(res, "2/udp") != 1 {
		t.Fail()
	}

	if strings.Contains(res, "3/tcp") || strings.Contains(res, "3/udp") {
		t.Fail()
	}
}

func TestStatCluster(t *testing.T) {
	t.Parallel()

	const rules = `[{"name": "foo", "if": "node.Listen.Has('1/tcp')"},
{"name": "bar", "if": "node.Listen.HasAny('2/tcp')"}]`

	cb := cluster.NewRules(builder.NewStat(), nil)

	if err := cb.FromReader(bytes.NewBufferString(rules)); err != nil {
		t.Fatal(err)
	}

	cfg := &graph.Config{
		Builder: cb,
		Meta:    &testEnricher{},
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	const raw = `{
   "name": "test1",
    "is_external": false,
    "networks": ["test"],
    "listen": {"1": ["1/udp", "1/tcp"]},
    "connected": {
        "test2":[
            {"src": "1", "dst": "2", "port": "2/tcp"},
            {"src": "1", "dst": "2", "port": "2/udp"}
        ],
        "test3":[
            {"src": "1", "dst": "3", "port": "3/tcp"}
        ]
      }
    }
    {
    "name": "test2",
    "is_external": false,
    "networks": ["test"],
    "listen": {"2":["2/tcp", "2/udp"]},
    "connected": {
        "test1":[
            {"src": "2", "dst": "1", "port": "1/tcp"},
            {"src": "2", "dst": "1", "port": "1/udp"}
        ],
        "test3": [
            {"src": "2", "dst": "3", "port": "3/udp"}
        ]
      }
    }`

	if err := ldr.FromReader(bytes.NewBufferString(raw)); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	var buf bytes.Buffer

	cb.Write(&buf)

	res := buf.String()

	if !strings.Contains(res, "Clusters") {
		t.Fail()
	}

	if strings.Count(res, "foo: 1") != 1 {
		t.Fail()
	}

	if strings.Count(res, "bar: 1") != 1 {
		t.Fail()
	}
}
