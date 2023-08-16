package node_test

import (
	"strings"
	"testing"

	"github.com/s0rg/decompose/internal/node"
)

func TestPortLabel(t *testing.T) {
	t.Parallel()

	p := node.Port{Value: 100, Kind: "tcp"}
	l := p.Label()

	if !strings.HasPrefix(l, p.String()) {
		t.Fail()
	}

	if !strings.HasSuffix(l, p.Kind) {
		t.Fail()
	}
}

func TestNodeIsExternal(t *testing.T) {
	t.Parallel()

	n := node.Node{}

	if !n.IsExternal() {
		t.Fail()
	}

	n.ID = "id"

	if n.IsExternal() {
		t.Fail()
	}
}
