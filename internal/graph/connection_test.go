package graph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/graph"
)

func TestConnectionIsListener(t *testing.T) {
	t.Parallel()

	c := graph.Connection{}

	if !c.IsListener() {
		t.Fail()
	}

	c.DstPort = 1

	if c.IsListener() {
		t.Fail()
	}
}

func TestConnectionIsInbound(t *testing.T) {
	t.Parallel()

	c := graph.Connection{}

	c.DstPort = 1

	if !c.IsInbound() {
		t.Fail()
	}

	c.SrcPort = 2

	if c.IsInbound() {
		t.Fail()
	}
}
