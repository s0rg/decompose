package cluster_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/cluster"
	"github.com/s0rg/decompose/internal/node"
	"github.com/s0rg/set"
)

func TestNodeMatch(t *testing.T) {
	t.Parallel()

	a := &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports:     []*node.Port{},
	}

	b := &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports:     []*node.Port{},
	}

	a.Inbounds.Add("1")
	a.Outbounds.Add("2")

	if a.Match("0", b) > 0.0 {
		t.FailNow()
	}

	if a.Match("1", b) == 0.0 {
		t.FailNow()
	}

	if a.Match("2", b) == 0.0 {
		t.FailNow()
	}
}
