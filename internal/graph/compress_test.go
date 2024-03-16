package graph_test

import (
	"errors"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

type testNamedBuilder struct {
	AddError   error
	WriteError error
	Nodes      int
	Edges      int
}

func (b *testNamedBuilder) AddNode(n *node.Node) error {
	log.Printf("%+v", n)

	b.Nodes++

	return b.AddError
}

func (b *testNamedBuilder) AddEdge(e *node.Edge) {
	log.Printf("%+v", e)

	b.Edges++
}

func (b *testNamedBuilder) Name() string {
	return "test-builder"
}

func (b *testNamedBuilder) Write(_ io.Writer) error {
	return b.WriteError
}

func TestCompressor(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}

	c := graph.NewCompressor(tb)

	// ingress
	nginx := &node.Node{
		ID:    "1",
		Name:  "nginx1",
		Ports: &node.Ports{},
	}

	nginx.Ports.Add("nginx", &node.Port{Kind: "tcp", Value: 443})
	c.AddNode(nginx)

	// apps
	app1 := &node.Node{
		ID:    "2",
		Name:  "app1",
		Ports: &node.Ports{},
	}

	app1.Ports.Add("app", &node.Port{Kind: "tcp", Value: 8080})
	c.AddNode(app1)

	app2 := &node.Node{
		ID:    "3",
		Name:  "app2",
		Ports: &node.Ports{},
	}

	app2.Ports.Add("app", &node.Port{Kind: "tcp", Value: 8080})
	c.AddNode(app2)

	// dbs
	bouncer := &node.Node{
		ID:    "4",
		Name:  "pgbouncer1",
		Ports: &node.Ports{},
	}

	bouncer.Ports.Add("pgbouncer", &node.Port{Kind: "tcp", Value: 5432})
	c.AddNode(bouncer)

	db1 := &node.Node{
		ID:    "5",
		Name:  "postgres1",
		Ports: &node.Ports{},
	}

	db1.Ports.Add("postgres", &node.Port{Kind: "tcp", Value: 5432})
	c.AddNode(db1)

	db2 := &node.Node{
		ID:    "6",
		Name:  "postgres2",
		Ports: &node.Ports{},
	}

	db2.Ports.Add("postgres", &node.Port{Kind: "tcp", Value: 5432})
	c.AddNode(db2)

	// external
	ext := &node.Node{
		ID:    "EXT",
		Name:  "EXT",
		Ports: &node.Ports{},
	}

	ext.Ports.Add("", &node.Port{Kind: "tcp", Value: 9000})
	c.AddNode(ext)

	// edges

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "2",
		Port:  &node.Port{Kind: "tcp", Value: 8080},
	})

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "3",
		Port:  &node.Port{Kind: "tcp", Value: 8080},
	})

	c.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "4",
		Port:  &node.Port{Kind: "tcp", Value: 5432},
	})

	c.AddEdge(&node.Edge{
		SrcID: "3",
		DstID: "4",
		Port:  &node.Port{Kind: "tcp", Value: 5432},
	})

	c.AddEdge(&node.Edge{
		SrcID: "4",
		DstID: "5",
		Port:  &node.Port{Kind: "tcp", Value: 5432},
	})

	c.AddEdge(&node.Edge{
		SrcID: "4",
		DstID: "5",
		Port:  &node.Port{Kind: "tcp", Value: 5432},
	})

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "2",
		Port:  &node.Port{Kind: "udp", Value: 22},
	})

	c.AddEdge(&node.Edge{
		SrcID: "X",
		DstID: "2",
		Port:  &node.Port{Kind: "udp", Value: 22},
	})

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "X",
		Port:  &node.Port{Kind: "udp", Value: 22},
	})

	c.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "EXT",
		Port:  &node.Port{Kind: "tcp", Value: 9000},
	})

	c.AddEdge(&node.Edge{
		SrcID: "3",
		DstID: "EXT",
		Port:  &node.Port{Kind: "tcp", Value: 9000},
	})

	if err := c.Write(nil); err != nil {
		t.Fail()
	}

	if tb.Edges != 5 || tb.Nodes != 5 {
		t.Fail()
	}
}

func TestCompressorName(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}

	c := graph.NewCompressor(tb)

	if !strings.Contains(c.Name(), "compressed") {
		t.Fail()
	}
}

func TestCompressorAddError(t *testing.T) {
	t.Parallel()

	myErr := errors.New("add error")
	tb := &testNamedBuilder{
		AddError: myErr,
	}

	c := graph.NewCompressor(tb)

	if err := c.AddNode(&node.Node{Ports: &node.Ports{}}); err != nil {
		t.Fail()
	}

	err := c.Write(nil)
	if !errors.Is(err, myErr) {
		t.Fail()
	}
}

func TestCompressorWriteError(t *testing.T) {
	t.Parallel()

	myErr := errors.New("write error")
	tb := &testNamedBuilder{
		WriteError: myErr,
	}

	c := graph.NewCompressor(tb)

	if err := c.AddNode(&node.Node{Ports: &node.Ports{}}); err != nil {
		t.Fail()
	}

	err := c.Write(nil)
	if !errors.Is(err, myErr) {
		t.Fail()
	}
}
