package cluster_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/cluster"
	"github.com/s0rg/decompose/internal/node"
)

func TestLayers(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{clusters: make(map[string][]string)}

	const similarity = 0.6

	ca := cluster.NewLayers(tb, similarity)

	if !strings.Contains(ca.Name(), strconv.FormatFloat(similarity, 'f', 1, 64)) {
		t.Fail()
	}

	_ = ca.AddNode(&node.Node{
		ID:   "6",
		Name: "node-6",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 6},
			{Kind: "tcp", Value: 1234},
			{Kind: "tcp", Value: 8080},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:   "1",
		Name: "node-1",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 80},
			{Kind: "tcp", Value: 443},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:   "2",
		Name: "node-2",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 2},
			{Kind: "tcp", Value: 1234},
			{Kind: "tcp", Value: 8080},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:   "3",
		Name: "node-3",
		Ports: []*node.Port{
			{Kind: "udp", Value: 53},
			{Kind: "tcp", Value: 8080},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:   "4",
		Name: "node-4",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 9090},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:   "5",
		Name: "node-5",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 1234},
			{Kind: "tcp", Value: 8081},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:   "R",
		Name: "R",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 22},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:   "5",
		Name: "node-5",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 1234},
			{Kind: "tcp", Value: 8081},
		},
	})

	_ = ca.AddNode(&node.Node{
		ID:    "10",
		Name:  "node-10",
		Ports: []*node.Port{},
	})

	ca.AddEdge("1", "2", &node.Port{Kind: "tcp", Value: 1234})
	ca.AddEdge("1", "3", &node.Port{Kind: "tcp", Value: 8080})
	ca.AddEdge("1", "6", &node.Port{Kind: "tcp", Value: 8080})
	ca.AddEdge("2", "4", &node.Port{Kind: "tcp", Value: 9090})
	ca.AddEdge("3", "4", &node.Port{Kind: "tcp", Value: 9090})
	ca.AddEdge("4", "3", &node.Port{Kind: "tcp", Value: 8080})
	ca.AddEdge("1", "5", &node.Port{Kind: "tcp", Value: 8081})
	ca.AddEdge("5", "4", &node.Port{Kind: "tcp", Value: 9090})
	ca.AddEdge("5", "R", &node.Port{Kind: "tcp", Value: 22})

	if tb.Nodes > 0 || tb.Edges > 0 {
		t.Fail()
	}

	if err := ca.Write(nil); err != nil {
		t.Fatal(err)
	}

	const (
		edgesDirect  = 9
		edgesCluster = 6

		wantNodes    = 8
		wantEdges    = edgesDirect + edgesCluster
		wantClusters = 4
	)

	if tb.Nodes != wantNodes || tb.Edges != wantEdges {
		t.Fail()
	}

	if tb.Clusters() != wantClusters {
		t.Fail()
	}
}

func TestLayersWriteError(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{
		Err: errors.New("test-error"),
	}

	const similarity = 0.5

	ca := cluster.NewLayers(tb, similarity)

	if err := ca.Write(nil); !errors.Is(err, tb.Err) {
		t.Fail()
	}
}

func TestLayersLabel(t *testing.T) {
	t.Parallel()

	s := []string{
		"foo-one1",
		"foo-two-1",
		"bar1",
		"bar-two-2",
		"barista",
		"doo",
		"doo2",
		"foo-1",
		"too-tee-1",
		"too-tee-2",
		"too-tee-3",
		"too-tee-4",
	}

	const (
		maxParts = 2
		want     = "too-foo"
	)

	if l := cluster.CreateLabel(s, 2); l != want {
		t.Log(l)
		t.Fail()
	}
}
