package graph_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

const (
	testBuilderName = "testbuilder"
	clusterRules    = `[{"name": "foo", "ports": ["80-80/tcp"]},
{"name": "bar", "ports": ["22/tcp", "443/tcp"]}]`
)

type testNamedBuilder struct {
	Err   error
	nodes int
	edges int
}

func (tb *testNamedBuilder) AddNode(_ *node.Node) error {
	if tb.Err != nil {
		return tb.Err
	}

	tb.nodes++

	return nil
}

func (tb *testNamedBuilder) AddEdge(_, _ string, _ node.Port) {
	tb.edges++
}

func (tb *testNamedBuilder) Name() string {
	return testBuilderName
}

func (tb *testNamedBuilder) Write(_ io.Writer) {}

func TestClusterError(t *testing.T) {
	t.Parallel()

	testCases := []string{
		`{`,
		`[{"name": "foo", "ports": "bar"}]`,
		`[{"name": "foo", "ports": ["#"]}]`,
		`[{"name": "foo", "ports": ["#/tcp"]}]`,
		`[{"name": "foo", "ports": ["-/tcp"]}]`,
		`[{"name": "foo", "ports": ["1-2-3/tcp"]}]`,
		`[{"name": "foo", "ports": ["1-/tcp"]}]`,
		`[{"name": "foo", "ports": ["-2/tcp"]}]`,
		`[{"name": "foo", "ports": ["10-1/tcp"]}]`,
		`[{"name": "foo", "ports": ["80/tcp"]}, {"name": "bar", "ports": ["70-90/tcp"]}]`,
	}

	ca := graph.NewClusterBuilder(nil)

	for _, tc := range testCases {
		if err := ca.FromReader(bytes.NewBufferString(tc)); err == nil {
			t.Fail()
		}
	}
}

func TestClusterMatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Node *node.Node
		Want string
	}{
		{
			Node: &node.Node{Ports: []node.Port{{Kind: "tcp", Value: 80}}},
			Want: "foo",
		},
		{
			Node: &node.Node{Ports: []node.Port{{Kind: "tcp", Value: 22}}},
			Want: "bar",
		},
		{
			Node: &node.Node{Ports: []node.Port{{Kind: "tcp", Value: 443}}},
			Want: "bar",
		},
		{
			Node: &node.Node{Ports: []node.Port{
				{Kind: "tcp", Value: 22},
				{Kind: "tcp", Value: 80},
				{Kind: "tcp", Value: 443},
			}},
			Want: "bar",
		},
		{
			Node: &node.Node{Ports: []node.Port{
				{Kind: "tcp", Value: 22},
				{Kind: "tcp", Value: 80},
				{Kind: "tcp", Value: 8080},
			}},
			Want: "foo",
		},
		{
			Node: &node.Node{Ports: []node.Port{
				{Kind: "sstp", Value: 5000},
			}},
			Want: "",
		},
	}

	ca := graph.NewClusterBuilder(nil)

	n := testCases[0].Node

	if _, ok := ca.Match(n); ok {
		t.Fail()
	}

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		m, ok := ca.Match(tc.Node)
		if tc.Want != "" && !ok {
			t.Fail()
		}

		if m != tc.Want {
			t.Fail()
		}
	}
}

func TestClusterMatchWeight(t *testing.T) {
	t.Parallel()

	const clusterRulesWeight = `[{"name": "foo", "ports": ["80-80/tcp"]},
{"name": "bar", "weight": 2, "ports": ["22/tcp", "443/tcp"]}]`

	testNode := &node.Node{Ports: []node.Port{
		{Kind: "tcp", Value: 22},
		{Kind: "tcp", Value: 80},
		{Kind: "tcp", Value: 8080},
	}}

	ca := graph.NewClusterBuilder(nil)

	if err := ca.FromReader(bytes.NewBufferString(clusterRulesWeight)); err != nil {
		t.Fatal(err)
	}

	m, ok := ca.Match(testNode)
	if !ok {
		t.Fail()
	}

	if m != "bar" {
		t.Fail()
	}
}

func TestClusterBuilder(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}
	ca := graph.NewClusterBuilder(tb)

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	ca.AddNode(&node.Node{
		ID: "1",
		Ports: []node.Port{
			{Kind: "tcp", Value: 80},
		}})

	ca.AddNode(&node.Node{
		ID: "2",
		Ports: []node.Port{
			{Kind: "tcp", Value: 22},
		}})

	ca.AddNode(&node.Node{
		ID: "3",
		Ports: []node.Port{
			{Kind: "tcp", Value: 443},
			{Kind: "tcp", Value: 8080},
		}})

	ca.AddNode(&node.Node{
		ID: "4",
		Ports: []node.Port{
			{Kind: "tcp", Value: 8080},
		}})

	ca.AddEdge("1", "3", node.Port{Kind: "tcp", Value: 443})
	ca.AddEdge("2", "3", node.Port{Kind: "tcp", Value: 8080})
	ca.AddEdge("3", "1", node.Port{Kind: "tcp", Value: 80})
	ca.AddEdge("1", "4", node.Port{Kind: "tcp", Value: 8080})
	ca.AddEdge("5", "1", node.Port{Kind: "tcp", Value: 80})
	ca.AddEdge("1", "5", node.Port{Kind: "tcp", Value: 80})

	if tb.edges != 2 || tb.nodes != 4 {
		t.Fail()
	}

	ca.Write(nil)

	if tb.edges != 4 {
		t.Fail()
	}
}

func TestClusterBuilderError(t *testing.T) {
	t.Parallel()

	myError := errors.New("test-error")

	tb := &testNamedBuilder{Err: myError}
	ca := graph.NewClusterBuilder(tb)

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	err := ca.AddNode(&node.Node{
		ID: "1",
		Ports: []node.Port{
			{Kind: "tcp", Value: 80},
		}})
	if !errors.Is(err, myError) {
		t.Fail()
	}
}

func TestClusterBuilderName(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}
	ca := graph.NewClusterBuilder(tb)

	name := ca.Name()

	if !strings.HasPrefix(name, testBuilderName) {
		t.Fail()
	}
}
