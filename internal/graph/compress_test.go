package graph_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

const diff = 3

type testNamedBuilder struct {
	AddError   error
	WriteError error
	Nodes      int
	Edges      int
}

func (b *testNamedBuilder) AddNode(_ *node.Node) error {
	b.Nodes++

	return b.AddError
}

func (b *testNamedBuilder) AddEdge(_ *node.Edge) {
	b.Edges++
}

func (b *testNamedBuilder) Name() string {
	return "test-builder"
}

func (b *testNamedBuilder) Write(_ io.Writer) error {
	return b.WriteError
}

func (b *testNamedBuilder) String() string {
	return fmt.Sprintf("nodes: %d edges: %d", b.Nodes, b.Edges)
}

func TestCompressor(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}

	c := graph.NewCompressor(tb, "", diff, false)

	// ingress
	nginx := &node.Node{
		ID:    "1",
		Name:  "nginx1",
		Ports: &node.Ports{},
	}

	nginx.Ports.Add("nginx", &node.Port{Kind: "tcp", Value: "443"})
	c.AddNode(nginx)

	// apps
	app1 := &node.Node{
		ID:    "2",
		Name:  "app1",
		Ports: &node.Ports{},
	}

	app1.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(app1)

	app2 := &node.Node{
		ID:    "3",
		Name:  "app2",
		Ports: &node.Ports{},
	}

	app2.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(app2)

	// dbs
	bouncer := &node.Node{
		ID:    "4",
		Name:  "pgbouncer1",
		Ports: &node.Ports{},
	}

	bouncer.Ports.Add("pgbouncer", &node.Port{Kind: "tcp", Value: "5432"})
	c.AddNode(bouncer)

	db1 := &node.Node{
		ID:    "5",
		Name:  "postgres1",
		Ports: &node.Ports{},
	}

	db1.Ports.Add("postgres", &node.Port{Kind: "tcp", Value: "5432"})
	c.AddNode(db1)

	db2 := &node.Node{
		ID:    "6",
		Name:  "postgres2",
		Ports: &node.Ports{},
	}

	db2.Ports.Add("postgres", &node.Port{Kind: "tcp", Value: "5432"})
	c.AddNode(db2)

	// external
	ext := &node.Node{
		ID:    "EXT",
		Name:  "EXT",
		Ports: &node.Ports{},
	}

	ext.Ports.Add("", &node.Port{Kind: "tcp", Value: "9000"})
	c.AddNode(ext)

	// edges

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "2",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "3",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "4",
		Port:  &node.Port{Kind: "tcp", Value: "5432"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "3",
		DstID: "4",
		Port:  &node.Port{Kind: "tcp", Value: "5432"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "4",
		DstID: "5",
		Port:  &node.Port{Kind: "tcp", Value: "5432"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "4",
		DstID: "5",
		Port:  &node.Port{Kind: "tcp", Value: "5432"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "2",
		Port:  &node.Port{Kind: "udp", Value: "22"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "X",
		DstID: "2",
		Port:  &node.Port{Kind: "udp", Value: "22"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "X",
		Port:  &node.Port{Kind: "udp", Value: "22"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "EXT",
		Port:  &node.Port{Kind: "tcp", Value: "9000"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "3",
		DstID: "EXT",
		Port:  &node.Port{Kind: "tcp", Value: "9000"},
	})

	if err := c.Write(nil); err != nil {
		t.Fail()
	}

	if tb.Edges != 5 || tb.Nodes != 4 {
		t.Fail()
	}
}

func TestCompressorForce(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}

	c := graph.NewCompressor(tb, "", 1, true)

	a1 := &node.Node{
		ID:    "a1-id",
		Name:  "a1",
		Ports: &node.Ports{},
	}

	a1.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(a1)

	a2 := &node.Node{
		ID:    "a2-id",
		Name:  "a2",
		Ports: &node.Ports{},
	}

	a2.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(a2)

	a3 := &node.Node{
		ID:    "a3-id",
		Name:  "a3",
		Ports: &node.Ports{},
	}

	a3.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(a3)

	b1 := &node.Node{
		ID:    "b1-id",
		Name:  "b1",
		Ports: &node.Ports{},
	}

	b1.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(b1)

	b2 := &node.Node{
		ID:    "b2-id",
		Name:  "b2",
		Ports: &node.Ports{},
	}

	b2.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(b2)

	b3 := &node.Node{
		ID:    "b3-id",
		Name:  "b3",
		Ports: &node.Ports{},
	}

	b3.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(b3)

	c1 := &node.Node{
		ID:    "c1-id",
		Name:  "c1",
		Ports: &node.Ports{},
	}

	c1.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(c1)

	c2 := &node.Node{
		ID:    "c2-id",
		Name:  "c2",
		Ports: &node.Ports{},
	}

	c2.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(c2)

	c3 := &node.Node{
		ID:    "c3-id",
		Name:  "c3",
		Ports: &node.Ports{},
	}

	c3.Ports.Add("app", &node.Port{Kind: "tcp", Value: "8080"})
	c.AddNode(c3)

	d1 := &node.Node{
		ID:    "d1-id",
		Name:  "d1",
		Ports: &node.Ports{},
	}

	d1.Ports.Add("app", &node.Port{Kind: "tcp", Value: "5555"})
	c.AddNode(d1)

	d2 := &node.Node{
		ID:    "d2-id",
		Name:  "d2",
		Ports: &node.Ports{},
	}

	d2.Ports.Add("app", &node.Port{Kind: "tcp", Value: "5555"})
	c.AddNode(d2)

	e1 := &node.Node{
		ID:    "e1-id",
		Name:  "e1",
		Ports: &node.Ports{},
	}

	e1.Ports.Add("app", &node.Port{Kind: "tcp", Value: "6666"})
	c.AddNode(e1)

	e2 := &node.Node{
		ID:    "e2-id",
		Name:  "e2",
		Ports: &node.Ports{},
	}

	e2.Ports.Add("app", &node.Port{Kind: "tcp", Value: "6666"})
	c.AddNode(e2)

	f1 := &node.Node{
		ID:    "f1-id",
		Name:  "f1",
		Ports: &node.Ports{},
	}

	f1.Ports.Add("app", &node.Port{Kind: "tcp", Value: "7777"})
	c.AddNode(f1)

	f2 := &node.Node{
		ID:    "f2-id",
		Name:  "f2",
		Ports: &node.Ports{},
	}

	f2.Ports.Add("app", &node.Port{Kind: "tcp", Value: "7777"})
	c.AddNode(f2)

	c.AddEdge(&node.Edge{
		SrcID: "a1-id",
		DstID: "b1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "a1-id",
		DstID: "c1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "a2-id",
		DstID: "b2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "a2-id",
		DstID: "c2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "b1-id",
		DstID: "a1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "b1-id",
		DstID: "c1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "b2-id",
		DstID: "a2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "b2-id",
		DstID: "c2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "c1-id",
		DstID: "a1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "c1-id",
		DstID: "b1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "c2-id",
		DstID: "a2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "c2-id",
		DstID: "a2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "a1-id",
		DstID: "d1-id",
		Port:  &node.Port{Kind: "tcp", Value: "5555"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "a2-id",
		DstID: "d2-id",
		Port:  &node.Port{Kind: "tcp", Value: "5555"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "b1-id",
		DstID: "e1-id",
		Port:  &node.Port{Kind: "tcp", Value: "6666"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "b2-id",
		DstID: "e2-id",
		Port:  &node.Port{Kind: "tcp", Value: "6666"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "e1-id",
		DstID: "b2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "e2-id",
		DstID: "b1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "f1-id",
		DstID: "c1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "f2-id",
		DstID: "c2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "a1-id",
		DstID: "a1-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	c.AddEdge(&node.Edge{
		SrcID: "a1-id",
		DstID: "a2-id",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	if err := c.Write(nil); err != nil {
		t.Fail()
	}

	if tb.Nodes != 3 || tb.Edges != 6 {
		t.Fail()
	}
}

func TestCompressorName(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}

	c := graph.NewCompressor(tb, "", diff, false)

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

	c := graph.NewCompressor(tb, "", diff, false)

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

	c := graph.NewCompressor(tb, "", diff, false)

	if err := c.AddNode(&node.Node{Ports: &node.Ports{}}); err != nil {
		t.Fail()
	}

	err := c.Write(nil)
	if !errors.Is(err, myErr) {
		t.Fail()
	}
}
