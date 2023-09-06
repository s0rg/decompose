package node_test

import (
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/node"
)

func TestPortLabelID(t *testing.T) {
	t.Parallel()

	const want = "100"

	p := node.Port{Value: 100, Kind: "tcp"}
	l := p.Label()

	if !strings.HasPrefix(l, want) {
		t.Fail()
	}

	if !strings.HasSuffix(l, p.Kind) {
		t.Fail()
	}

	id := p.ID()

	if !strings.HasSuffix(id, want) {
		t.Fail()
	}

	if !strings.HasPrefix(id, p.Kind) {
		t.Fail()
	}
}

func TestPortsDedup(t *testing.T) {
	t.Parallel()

	ports := []*node.Port{
		{Kind: "tcp", Value: 1},
		{Kind: "udp", Value: 1},
		{Kind: "tcp", Value: 1},
		{Kind: "udp", Value: 2},
		{Kind: "tcp", Value: 3},
	}

	rv := node.Ports(ports).Dedup()

	if len(rv) != 4 {
		t.Fail()
	}
}

func TestPortsHas(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Ports  node.Ports
		Labels []string
		Want   bool
	}{
		{
			Ports:  []*node.Port{},
			Labels: []string{},
			Want:   false,
		},
		{
			Ports:  []*node.Port{},
			Labels: []string{"80/tcp"},
			Want:   false,
		},
		{
			Ports: []*node.Port{
				{Kind: "tcp", Value: 80},
			},
			Labels: []string{"80/tcp"},
			Want:   true,
		},
		{
			Ports: []*node.Port{
				{Kind: "tcp", Value: 81},
			},
			Labels: []string{"80/tcp"},
			Want:   false,
		},
		{
			Ports: []*node.Port{
				{Kind: "tcp", Value: 80},
			},
			Labels: []string{"80/tcp", "443/tcp"},
			Want:   false,
		},
	}

	for _, tc := range testCases {
		if tc.Ports.Has(tc.Labels...) != tc.Want {
			t.Fail()
		}
	}
}

func TestPortsHasAny(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Ports  node.Ports
		Labels []string
		Want   bool
	}{
		{
			Ports:  []*node.Port{},
			Labels: []string{},
			Want:   false,
		},
		{
			Ports:  []*node.Port{},
			Labels: []string{"80/tcp"},
			Want:   false,
		},
		{
			Ports: []*node.Port{
				{Kind: "tcp", Value: 80},
			},
			Labels: []string{"80/tcp"},
			Want:   true,
		},
		{
			Ports: []*node.Port{
				{Kind: "tcp", Value: 81},
			},
			Labels: []string{"80/tcp"},
			Want:   false,
		},
		{
			Ports: []*node.Port{
				{Kind: "tcp", Value: 80},
			},
			Labels: []string{"80/tcp", "443/tcp"},
			Want:   true,
		},
	}

	for _, tc := range testCases {
		if tc.Ports.HasAny(tc.Labels...) != tc.Want {
			t.Fail()
		}
	}
}
