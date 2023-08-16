package builder

import (
	"io"

	"github.com/s0rg/decompose/internal/netgraph"
)

const (
	kindDOT  = "dot"
	kindJSON = "json"
)

type Builder interface {
	netgraph.Builder
	Write(io.Writer)
}

func Create(kind string) (b Builder, ok bool) {
	switch kind {
	case kindDOT:
		return NewDOT(), true
	case kindJSON:
		return NewJSON(), true
	}

	return
}
