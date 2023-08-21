package graph_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

func TestExtraFromReaderErr(t *testing.T) {
	t.Parallel()

	r := bytes.NewBufferString(`{`)
	l := graph.NewMetaLoader()

	if err := l.FromReader(r); err == nil {
		t.Fatal("err nil")
	}
}

func TestExtraEnrich(t *testing.T) {
	t.Parallel()

	r := bytes.NewBufferString(`
{
    "foo": {
        "info": "its a foo",
        "tags": ["foo"]
    },
    "baz": {
        "info": "its a baz",
        "tags": ["not-foo", "baz"]
    }
}
`)

	testCases := []struct {
		Node        node.Node
		WantInfoKey string
		WantTagsNum int
		Want        bool
	}{
		{
			Node:        node.Node{Name: "foo-1"},
			Want:        true,
			WantInfoKey: "foo",
			WantTagsNum: 1,
		},
		{
			Node:        node.Node{Name: "baz-2"},
			Want:        true,
			WantInfoKey: "baz",
			WantTagsNum: 2,
		},
		{
			Node: node.Node{Name: "bar-1"},
		},
	}

	l := graph.NewMetaLoader()

	// empty state test

	n := testCases[0].Node

	l.Enrich(&n)

	if n.Meta != nil {
		t.Fatal("not nill")
	}

	if err := l.FromReader(r); err != nil {
		t.Fatal("reader err=", err)
	}

	for i, tc := range testCases {
		n := tc.Node

		l.Enrich(&n)

		if tc.Want && n.Meta == nil {
			t.Fatalf("case[%d] state", i)
		}

		if !tc.Want {
			continue
		}

		if !strings.Contains(n.Meta.Info, tc.WantInfoKey) {
			t.Fatalf("case[%d] extra key", i)
		}

		if len(n.Meta.Tags) != tc.WantTagsNum {
			t.Fatalf("case[%d] tags", i)
		}
	}
}
