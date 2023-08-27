//go:build !test

package builder

import (
	"io"

	"github.com/s0rg/decompose/internal/graph"
)

const (
	kindDOT         = "dot"
	kindJSON        = "json"
	kindTREE        = "tree"
	kindStructurizr = "sdsl"
)

type Builder interface {
	graph.Builder
	Write(io.Writer)
	Name() string
}

func Create(kind string) (b Builder, ok bool) {
	switch kind {
	case kindDOT:
		return NewDOT(), true
	case kindJSON:
		return NewJSON(), true
	case kindStructurizr:
		return NewStructurizr(), true
	case kindTREE:
		return NewTree(), true
	}

	return
}
