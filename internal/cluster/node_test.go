package cluster_test

import (
	"testing"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/cluster"
	"github.com/s0rg/decompose/internal/node"
)

func TestNodeMatch(t *testing.T) {
	t.Parallel()

	a := &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports:     &node.Ports{},
	}

	b := &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports:     &node.Ports{},
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

func TestNodeMatchPorts(t *testing.T) {
	t.Parallel()

	a := &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports:     &node.Ports{},
	}

	a.Ports.Add("", &node.Port{Kind: "tcp", Value: 1})

	b := &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports:     &node.Ports{},
	}

	b.Ports.Add("", &node.Port{Kind: "tcp", Value: 1})
	b.Ports.Add("", &node.Port{Kind: "tcp", Value: 5})

	a.Inbounds.Add("1")
	a.Outbounds.Add("2")

	if a.Match("", b) != 0.5 {
		t.Fail()
	}
}
