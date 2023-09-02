package graph_test

import (
	"bytes"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

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

	ca := graph.NewClusterAssigner()

	for _, tc := range testCases {
		if err := ca.FromReader(bytes.NewBufferString(tc)); err == nil {
			t.Fail()
		}
	}
}

func TestClusterAssign(t *testing.T) {
	t.Parallel()

	const rules = `[{"name": "foo", "ports": ["80-80/tcp"]},
    {"name": "bar", "ports": ["22/tcp", "443/tcp"]}]`

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

	ca := graph.NewClusterAssigner()

	n := testCases[0].Node

	ca.Assign(n)

	if n.Cluster != "" {
		t.Fail()
	}

	if err := ca.FromReader(bytes.NewBufferString(rules)); err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		n := tc.Node

		ca.Assign(n)

		if n.Cluster != tc.Want {
			t.Fail()
		}
	}
}
