package node_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/node"
)

func TestPortsGet(t *testing.T) {
	t.Parallel()

	ps := &node.Ports{}

	if ps.Len() > 0 {
		t.Fail()
	}

	port := &node.Port{Kind: "tcp", Value: "1"}

	ps.Add("foo", port)

	if ps.Len() != 1 {
		t.Fail()
	}

	val, ok := ps.Get(port)
	if !ok || val != "foo" {
		t.Fail()
	}

	if _, ok = ps.Get(&node.Port{}); ok {
		t.Fail()
	}
}
