package netgraph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/netgraph"
)

func TestConnectionIsListener(t *testing.T) {
	t.Parallel()

	c := netgraph.Connection{}

	if !c.IsListener() {
		t.Fail()
	}

	c.RemotePort = 1

	if c.IsListener() {
		t.Fail()
	}
}

func TestConnectionIsInbound(t *testing.T) {
	t.Parallel()

	c := netgraph.Connection{}

	c.RemotePort = 1

	if !c.IsInbound() {
		t.Fail()
	}

	c.LocalPort = 2

	if c.IsInbound() {
		t.Fail()
	}
}
