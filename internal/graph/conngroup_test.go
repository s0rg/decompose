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

	cg.AddListener(&graph.Connection{Proto: graph.TCP, SrcPort: 1})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, SrcPort: 1})

	cg.AddListener(&graph.Connection{Proto: graph.TCP, SrcPort: 1}) // duplicate
	cg.AddListener(&graph.Connection{Proto: graph.TCP, DstPort: 1}) // invalid

	if cg.Len() != 2 {
		t.Fail()
	}

	cg.Sort()

	cg.IterOutbounds(func(_ *graph.Connection) {
		t.Fail()
	})

	cg.IterListeners(func(c *graph.Connection) {
		if c.SrcPort != 1 {
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

	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, SrcPort: 2, DstPort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.UDP, SrcPort: 3, DstPort: 1})

	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, SrcPort: 2, DstPort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, DstPort: 1})

	if cg.Len() != 2 {
		t.Fail()
	}

	cg.Sort()

	cg.IterListeners(func(_ *graph.Connection) {
		t.Fail()
	})

	cg.IterOutbounds(func(c *graph.Connection) {
		if c.DstPort != 1 {
			t.Fail()
		}
	})
}

func TestConnGroupSort(t *testing.T) {
	t.Parallel()

	cg := &graph.ConnGroup{}

	cg.AddListener(&graph.Connection{Proto: graph.TCP, SrcPort: 1, Listen: true})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, SrcPort: 1, Listen: true})
	cg.AddListener(&graph.Connection{Proto: graph.TCP, SrcPort: 2, Listen: true})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, SrcPort: 2, Listen: true})
	cg.AddListener(&graph.Connection{Proto: graph.TCP, SrcPort: 3, Listen: true})
	cg.AddListener(&graph.Connection{Proto: graph.UDP, SrcPort: 3, Listen: true})

	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, SrcPort: 2, DstPort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.UDP, SrcPort: 3, DstPort: 1})
	cg.AddOutbound(&graph.Connection{Proto: graph.TCP, SrcPort: 4, DstPort: 2})
	cg.AddOutbound(&graph.Connection{Proto: graph.UDP, SrcPort: 5, DstPort: 2})

	cg.Sort()

	if cg.Len() != 10 {
		t.Fail()
	}
}
