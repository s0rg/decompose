package cluster_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/cluster"
	"github.com/s0rg/decompose/internal/node"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	g := cluster.NewGrouper(0.9)

	g.Add("1", []*node.Port{
		{Kind: "tcp", Value: 1},
		{Kind: "tcp", Value: 2},
	})
	g.Add("2", []*node.Port{
		{Kind: "tcp", Value: 3},
		{Kind: "tcp", Value: 4},
	})
	g.Add("3", []*node.Port{
		{Kind: "tcp", Value: 1},
		{Kind: "tcp", Value: 3},
	})
	g.Add("4", []*node.Port{
		{Kind: "tcp", Value: 1},
		{Kind: "tcp", Value: 2},
	})
	g.Add("5", []*node.Port{
		{Kind: "tcp", Value: 1},
		{Kind: "tcp", Value: 4},
	})

	var groups int

	g.IterGroups(func(_ int, _ []string) {
		groups++
	})

	if groups != 2 {
		t.Fail()
	}
}
