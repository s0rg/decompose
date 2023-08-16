package netgraph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/netgraph"
)

func TestParseNetproto(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Val   string
		Valid bool
		Want  netgraph.NetProto
	}{
		{Val: "tcp", Valid: true, Want: netgraph.TCP},
		{Val: "udp", Valid: true, Want: netgraph.UDP},
		{Val: "all", Valid: true, Want: netgraph.ALL},
		{Val: "bad", Valid: false},
	}

	for i, tc := range testCases {
		got, ok := netgraph.ParseNetProto(tc.Val)

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
		{Val: "all", Want: "tu"},
	}

	for i, tc := range testCases {
		p, _ := netgraph.ParseNetProto(tc.Val)

		if p.Flag() != tc.Want {
			t.Fatalf("case[%d] failed for '%s' want: %s got: %s", i, tc.Val, tc.Want, p.Flag())
		}
	}
}
