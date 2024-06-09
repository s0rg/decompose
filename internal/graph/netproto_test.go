package graph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/graph"
)

func TestParseNetproto(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Val   string
		Valid bool
		Want  graph.NetProto
	}{
		{Val: "tcp", Valid: true, Want: graph.TCP},
		{Val: "udp", Valid: true, Want: graph.UDP},
		{Val: "all", Valid: true, Want: graph.ALL},
		{Val: "unix", Valid: true, Want: graph.UNIX},
		{Val: "bad", Valid: false},
	}

	for i, tc := range testCases {
		got, ok := graph.ParseNetProto(tc.Val)

		if ok != tc.Valid {
			t.Fatalf("case[%d] failed for '%s' want: %t got: %t", i, tc.Val, tc.Valid, ok)
		}

		if !ok {
			continue
		}

		if got != tc.Want {
			t.Fatalf("case[%d] failed for '%s' want: %s got: %s", i, tc.Val, tc.Want.String(), got.String())
		}
	}
}

func TestNetprotoString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Val  string
		Want string
	}{
		{Val: "tcp", Want: "t"},
		{Val: "udp", Want: "u"},
		{Val: "unix", Want: "x"},
		{Val: "all", Want: "tux"},
	}

	for i, tc := range testCases {
		p, _ := graph.ParseNetProto(tc.Val)

		if p.Flag() != tc.Want {
			t.Fatalf("case[%d] failed for '%s' want: %s got: %s", i, tc.Val, tc.Want, p.Flag())
		}
	}
}
