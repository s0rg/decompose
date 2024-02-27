package graph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/graph"
)

func TestConnGroupListeners(t *testing.T) {
	t.Parallel()

	cg := &graph.ConnGroup{}

	if cg.Len() > 0 {
		t.Fail()
	}

	cg.AddListener(&graph.Connection{Proto: graph.TCP, LocalPort: 1})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, LocalPort: 1})

	cg.AddListener(&graph.Connection{Proto: graph.TCP, LocalPort: 1})  // duplicate
	cg.AddListener(&graph.Connection{Proto: graph.TCP, RemotePort: 1}) // invalid

	if cg.Len() != 2 {
		t.Fail()
	}

	cg.Sort()

	cg.IterOutbounds(func(_ *graph.Connection) {
		t.Fail()
	})

	cg.IterListeners(func(c *graph.Connection) {
		if c.LocalPort != 1 {
			t.Fail()
		}
	})
}

func TestConnGroupOutbounds(t *testing.T) {
	t.Parallel()

	cg := &graph.ConnGroup{}

	if cg.Len() > 0 {
		t.Fail()
	}

	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, LocalPort: 2, RemotePort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.UDP, LocalPort: 3, RemotePort: 1})

	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, LocalPort: 2, RemotePort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, RemotePort: 1})

	if cg.Len() != 2 {
		t.Fail()
	}

	cg.Sort()

	cg.IterListeners(func(_ *graph.Connection) {
		t.Fail()
	})

	cg.IterOutbounds(func(c *graph.Connection) {
		if c.RemotePort != 1 {
			t.Fail()
		}
	})
}

func TestConnGroupSort(t *testing.T) {
	t.Parallel()

	cg := &graph.ConnGroup{}

	cg.AddListener(&graph.Connection{Proto: graph.TCP, LocalPort: 1})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, LocalPort: 1})
	cg.AddListener(&graph.Connection{Proto: graph.TCP, LocalPort: 2})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, LocalPort: 2})
	cg.AddListener(&graph.Connection{Proto: graph.TCP, LocalPort: 3})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, LocalPort: 3})

	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, LocalPort: 2, RemotePort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.UDP, LocalPort: 3, RemotePort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, LocalPort: 4, RemotePort: 2})
	cg.AddOutbound(&graph.Connection{Proto: graph.UDP, LocalPort: 5, RemotePort: 2})

	cg.Sort()

	if cg.Len() != 10 {
		t.Fail()
	}
}
