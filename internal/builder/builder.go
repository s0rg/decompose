//go:build !test

package builder

import (
	"github.com/s0rg/decompose/internal/graph"
)

const (
	KindDOT         = "dot"
	KindJSON        = "json"
	KindTREE        = "tree"
	KindYAML        = "yaml"
	KindSTAT        = "stat"
	KindStructurizr = "sdsl"
)

func Create(kind string) (b graph.NamedBuilderWriter, ok bool) {
	switch kind {
	case KindDOT:
		return NewDOT(), true
	case KindJSON:
		return NewJSON(), true
	case KindStructurizr:
		return NewStructurizr(), true
	case KindTREE:
		return NewTree(), true
	case KindYAML:
		return NewYAML(), true
	case KindSTAT:
		return NewStat(), true
	}

	return
}

func SupportCluster(n string) (yes bool) {
	switch n {
	case KindDOT, KindStructurizr, KindSTAT:
		return true
	}

	return false
}

func Names() (rv []string) {
	return []string{
		KindDOT,
		KindJSON,
		KindTREE,
		KindYAML,
		KindSTAT,
		KindStructurizr,
	}
}
