package node_test

import (
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/node"
)

func makeTestPorts(vals ...*node.Port) (rv *node.Ports) {
	rv = &node.Ports{}

	for _, p := range vals {
		rv.Add("", p)
	}

	return rv
}

func TestPortLabel(t *testing.T) {
	t.Parallel()

	const want = "100"

	p := node.Port{Value: "100", Kind: "tcp"}
	l := p.Label()

	if !strings.HasPrefix(l, p.Kind) {
		t.Fail()
	}

	if !strings.HasSuffix(l, want) {
		t.Fail()
	}
}

func TestPortsHas(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Ports  *node.Ports
		Labels []string
		Want   bool
	}{
		{
			Ports:  &node.Ports{},
			Labels: []string{},
			Want:   false,
		},
		{
			Ports:  &node.Ports{},
			Labels: []string{"tcp:80"},
			Want:   false,
		},
		{
			Ports: makeTestPorts(&node.Port{
				Kind:  "tcp",
				Value: "80",
			}),
			Labels: []string{"tcp:80"},
			Want:   true,
		},
		{
			Ports: makeTestPorts(&node.Port{
				Kind:  "tcp",
				Value: "81",
			}),
			Labels: []string{"tcp:80"},
			Want:   false,
		},
		{
			Ports: makeTestPorts(&node.Port{
				Kind:  "tcp",
				Value: "80",
			}),
			Labels: []string{"tcp:80", "tcp:443"},
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
		Ports  *node.Ports
		Labels []string
		Want   bool
	}{
		{
			Ports:  &node.Ports{},
			Labels: []string{},
			Want:   false,
		},
		{
			Ports:  &node.Ports{},
			Labels: []string{"tcp:80"},
			Want:   false,
		},
		{
			Ports: makeTestPorts(&node.Port{
				Kind:  "tcp",
				Value: "80",
			}),
			Labels: []string{"tcp:80"},
			Want:   true,
		},
		{
			Ports: makeTestPorts(&node.Port{
				Kind:  "tcp",
				Value: "81",
			}),
			Labels: []string{"tcp:80"},
			Want:   false,
		},
		{
			Ports: makeTestPorts(&node.Port{
				Kind:  "tcp",
				Value: "80",
			}),
			Labels: []string{"tcp:80", "tcp:443"},
			Want:   true,
		},
	}

	for _, tc := range testCases {
		if tc.Ports.HasAny(tc.Labels...) != tc.Want {
			t.Fail()
		}
	}
}
