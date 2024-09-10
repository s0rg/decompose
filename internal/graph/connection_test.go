package graph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/graph"
)

func TestConnectionIsListener(t *testing.T) {
	t.Parallel()

	c := graph.Connection{}

	if c.IsListener() {
		t.Fail()
	}

	c.Listen = true

	if !c.IsListener() {
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

func TestConnectionUNIX(t *testing.T) {
	t.Parallel()

	const uniqID = "/some/unix.sock"

	c := graph.Connection{
		Proto: graph.UNIX,
		Path:  uniqID,
	}

	if c.IsLocal() {
		t.Fail()
	}

	if c.IsInbound() {
		t.Fail()
	}

	c.Listen = true

	if !c.IsInbound() {
		t.Fail()
	}

	if _, ok := c.UniqID(); !ok {
		t.Fail()
	}
}
