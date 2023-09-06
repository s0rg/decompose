//go:build !test

package builder

import (
	"github.com/s0rg/decompose/internal/graph"
)

const (
	kindDOT         = "dot"
	kindJSON        = "json"
	kindTREE        = "tree"
	kindYAML        = "yaml"
	kindSTAT        = "stat"
	kindStructurizr = "sdsl"
)

func Create(kind string) (b graph.NamedBuilderWriter, ok bool) {
	switch kind {
	case kindDOT:
		return NewDOT(), true
	case kindJSON:
		return NewJSON(), true
	case kindStructurizr:
		return NewStructurizr(), true
	case kindTREE:
		return NewTree(), true
	case kindYAML:
		return NewYAML(), true
	case kindSTAT:
		return NewStat(), true
	}

	return
}

func SupportCluster(n string) (yes bool) {
	return n == kindDOT || n == kindStructurizr
}

func Names() (rv []string) {
	return []string{
		kindDOT,
		kindJSON,
		kindTREE,
		kindYAML,
		kindSTAT,
		kindStructurizr,
	}
}
