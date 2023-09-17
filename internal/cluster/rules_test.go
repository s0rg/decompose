package cluster_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/antonmedv/expr/vm"

	"github.com/s0rg/decompose/internal/cluster"
	"github.com/s0rg/decompose/internal/node"
)

const (
	testBuilderName = "testbuilder"
	clusterRules    = `[{"name": "foo", "if": "node.Listen.Has('80/tcp')"},
{"name": "bar", "if": "node.Listen.HasAny('22/tcp', '443/tcp')"}]`
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

func (tb *testNamedBuilder) AddEdge(_, _ string, _ *node.Port) {
	tb.edges++
}

func (tb *testNamedBuilder) Name() string            { return testBuilderName }
func (tb *testNamedBuilder) Write(_ io.Writer) error { return tb.Err }

func TestClusterError(t *testing.T) {
	t.Parallel()

	testCases := []string{
		`{`,
		`[{"name": "foo", "if": ""}]`,
		`[{"name": "foo", "if": "#"}]`,
	}

	ca := cluster.NewRules(nil, nil)

	for _, tc := range testCases {
		if err := ca.FromReader(bytes.NewBufferString(tc)); err == nil {
			t.Fail()
		}
	}
}

func TestRulesMatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Node *node.Node
		Want string
	}{
		{
			Node: &node.Node{Ports: []*node.Port{{Kind: "tcp", Value: 80}}},
			Want: "foo",
		},
		{
			Node: &node.Node{Ports: []*node.Port{{Kind: "tcp", Value: 22}}},
			Want: "bar",
		},
		{
			Node: &node.Node{Ports: []*node.Port{{Kind: "tcp", Value: 443}}},
			Want: "bar",
		},
		{
			Node: &node.Node{Ports: []*node.Port{
				{Kind: "tcp", Value: 22},
				{Kind: "tcp", Value: 80},
				{Kind: "tcp", Value: 443},
			}},
			Want: "foo",
		},
		{
			Node: &node.Node{Ports: []*node.Port{
				{Kind: "tcp", Value: 22},
				{Kind: "tcp", Value: 80},
				{Kind: "tcp", Value: 8080},
			}},
			Want: "foo",
		},
		{
			Node: &node.Node{Ports: []*node.Port{
				{Kind: "sstp", Value: 5000},
			}},
			Want: "",
		},
	}

	ca := cluster.NewRules(nil, nil)

	n := testCases[0].Node

	if _, ok := ca.Match(n); ok {
		t.Fail()
	}

	if ca.CountRules() != 0 {
		t.Fail()
	}

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	if ca.CountRules() != 2 {
		t.Fail()
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

func TestRulesMatchWeight(t *testing.T) {
	t.Parallel()

	const clusterRulesWeight = `[{"name": "foo", "weight": 2, "if": "node.Listen.Has('80/tcp')"},
{"name": "bar", "if": "node.Listen.HasAny('22/tcp', '443/tcp')"}]`

	testNode := &node.Node{Ports: []*node.Port{
		{Kind: "tcp", Value: 22},
		{Kind: "tcp", Value: 80},
		{Kind: "tcp", Value: 8080},
	}}

	ca := cluster.NewRules(nil, nil)

	if err := ca.FromReader(bytes.NewBufferString(clusterRulesWeight)); err != nil {
		t.Fatal(err)
	}

	m, ok := ca.Match(testNode)
	if !ok {
		t.Fail()
	}

	if m != "foo" {
		t.Fail()
	}
}

func TestRules(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}
	ca := cluster.NewRules(tb, nil)

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	ca.AddNode(&node.Node{
		ID: "1",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 80},
		}})

	ca.AddNode(&node.Node{
		ID: "2",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 22},
		}})

	ca.AddNode(&node.Node{
		ID: "3",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 443},
			{Kind: "tcp", Value: 8080},
		}})

	ca.AddNode(&node.Node{
		ID: "4",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 8080},
		}})

	ca.AddEdge("1", "3", &node.Port{Kind: "tcp", Value: 443})
	ca.AddEdge("2", "3", &node.Port{Kind: "tcp", Value: 8080})
	ca.AddEdge("3", "1", &node.Port{Kind: "tcp", Value: 80})
	ca.AddEdge("1", "4", &node.Port{Kind: "tcp", Value: 8080})
	ca.AddEdge("5", "1", &node.Port{Kind: "tcp", Value: 80})
	ca.AddEdge("1", "5", &node.Port{Kind: "tcp", Value: 80})

	if tb.edges != 4 || tb.nodes != 4 {
		t.Fail()
	}

	ca.Write(nil)

	if tb.edges != 7 {
		t.Fail()
	}
}

func TestRulesMatchError(t *testing.T) {
	t.Parallel()

	myError := errors.New("test-error")

	tb := &testNamedBuilder{}
	ca := cluster.NewRules(tb, func(_ *vm.Program, _ any) (any, error) {
		return nil, myError
	})

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	if _, ok := ca.Match(&node.Node{}); ok {
		t.Fail()
	}
}

func TestRulesBuilderAddError(t *testing.T) {
	t.Parallel()

	myError := errors.New("test-error")

	tb := &testNamedBuilder{Err: myError}
	ca := cluster.NewRules(tb, nil)

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	err := ca.AddNode(&node.Node{
		ID: "1",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 80},
		}})
	if !errors.Is(err, myError) {
		t.Fail()
	}
}

func TestRulesBuilderName(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}
	ca := cluster.NewRules(tb, nil)

	name := ca.Name()

	if !strings.HasPrefix(name, testBuilderName) {
		t.Fail()
	}
}

func TestRulesBuilderWriteError(t *testing.T) {
	t.Parallel()

	tb := &testNamedBuilder{}
	ca := cluster.NewRules(tb, nil)

	if err := ca.FromReader(bytes.NewBufferString(clusterRules)); err != nil {
		t.Fatal(err)
	}

	ca.AddNode(&node.Node{
		ID: "1",
		Ports: []*node.Port{
			{Kind: "tcp", Value: 80},
		}})

	tb.Err = errors.New("test-error")

	if err := ca.Write(nil); !errors.Is(err, tb.Err) {
		t.Fail()
	}
}
