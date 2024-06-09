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

	ca := cluster.NewLayers(tb, similarity, "foo")

	if !strings.Contains(ca.Name(), strconv.FormatFloat(similarity, 'f', 1, 64)) {
		t.Fail()
	}

	_ = ca.AddNode(&node.Node{
		ID:   "6",
		Name: "node-6",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "6"},
			{Kind: "tcp", Value: "1234"},
			{Kind: "tcp", Value: "8080"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:   "1",
		Name: "node-1",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "80"},
			{Kind: "tcp", Value: "443"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:   "2",
		Name: "node-2",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "2"},
			{Kind: "tcp", Value: "1234"},
			{Kind: "tcp", Value: "8080"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:   "3",
		Name: "node-3",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "udp", Value: "53"},
			{Kind: "tcp", Value: "8080"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:   "4",
		Name: "node-4",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "9090"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:   "5",
		Name: "node-5",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1234"},
			{Kind: "tcp", Value: "8081"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:   "R",
		Name: "R",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "22"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:   "5",
		Name: "node-5",
		Ports: makeTestPorts([]*node.Port{
			{Kind: "tcp", Value: "1234"},
			{Kind: "tcp", Value: "8081"},
		}),
	})

	_ = ca.AddNode(&node.Node{
		ID:    "10",
		Name:  "node-10",
		Ports: &node.Ports{},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "2",
		Port:  &node.Port{Kind: "tcp", Value: "1234"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "3",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "6",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "2",
		DstID: "4",
		Port:  &node.Port{Kind: "tcp", Value: "9090"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "3",
		DstID: "2",
		Port:  &node.Port{Kind: "tcp", Value: "9090"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "4",
		DstID: "3",
		Port:  &node.Port{Kind: "tcp", Value: "8080"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "1",
		DstID: "5",
		Port:  &node.Port{Kind: "tcp", Value: "8081"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "5",
		DstID: "4",
		Port:  &node.Port{Kind: "tcp", Value: "9090"},
	})

	ca.AddEdge(&node.Edge{
		SrcID: "5",
		DstID: "R",
		Port:  &node.Port{Kind: "tcp", Value: "22"},
	})

	if err := ca.Write(nil); err != nil {
		t.Fatal(err)
	}

	const (
		edgesDirect  = 9
		edgesCluster = 8

		wantNodes    = 8
		wantEdges    = edgesDirect + edgesCluster
		wantClusters = 6
	)

	if tb.Nodes != wantNodes || tb.Edges != wantEdges {
		t.Log("nodes:", tb.Nodes, "want:", wantNodes, "edges:", tb.Edges, "want:", wantEdges)
		t.Fail()
	}

	if tb.Clusters() != wantClusters {
		t.Log("clusters:", tb.Clusters(), "want:", wantClusters)
		t.Fail()
	}
}

func TestLayersWriteError(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{
		Err: errors.New("test-error"),
	}

	const similarity = 0.5

	ca := cluster.NewLayers(tb, similarity, "")

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
		want1    = "too-foo"
		want2    = "too-bar"
	)

	switch l := cluster.CreateLabel(s, 2); l {
	case want1, want2:
	default:
		t.Log(l)
		t.Fail()
	}
}
