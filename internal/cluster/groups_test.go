package cluster_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/cluster"
	"github.com/s0rg/decompose/internal/node"
	"github.com/s0rg/set"
)

func makeTestPorts(vals []*node.Port) (rv *node.Ports) {
	rv = &node.Ports{}

	for _, p := range vals {
		rv.Add("", p)
	}

	return rv
}

func TestAdd(t *testing.T) {
	t.Parallel()

	g := cluster.NewGrouper(0.5)

	g.Add("1", &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1", Number: 1},
			{Kind: "tcp", Value: "2", Number: 2},
			{Kind: "tcp", Value: "3", Number: 3},
		}),
	})

	g.Add("11", &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1", Number: 1},
			{Kind: "tcp", Value: "2", Number: 2},
			{Kind: "tcp", Value: "3", Number: 3},
		}),
	})

	g.Add("2", &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "3", Number: 3},
			{Kind: "tcp", Value: "2", Number: 2},
		}),
	})

	g.Add("3", &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1", Number: 1},
			{Kind: "tcp", Value: "3", Number: 3},
		}),
	})

	g.Add("4", &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1", Number: 1},
			{Kind: "tcp", Value: "2", Number: 2},
		}),
	})

	g.Add("5", &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1", Number: 1},
			{Kind: "tcp", Value: "2", Number: 2},
			{Kind: "tcp", Value: "4", Number: 4},
		}),
	})

	g.Add("6", &cluster.Node{
		Inbounds:  make(set.Unordered[string]),
		Outbounds: make(set.Unordered[string]),
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1", Number: 1},
		}),
	})

	var groups int

	g.IterGroups(func(_ int, _ []string) {
		groups++
	})

	if groups != 2 {
		t.Log(groups)
		t.Fail()
	}
}
