package builder

import (
	"github.com/s0rg/decompose/internal/graph"
)

const (
	KindCSV         = "csv"
	KindDOT         = "dot"
	KindJSON        = "json"
	KindTREE        = "tree"
	KindYAML        = "yaml"
	KindSTAT        = "stat"
	KindStructurizr = "sdsl"
	KindPlantUML    = "puml"
)

var Names = []string{
	KindCSV,
	KindDOT,
	KindJSON,
	KindTREE,
	KindYAML,
	KindSTAT,
	KindStructurizr,
	KindPlantUML,
}

func Create(kind string) (b graph.NamedBuilderWriter, ok bool) {
	switch kind {
	case KindCSV:
		return NewCSV(), true
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
	case KindPlantUML:
		return NewPlantUML(), true
	}

	return
}

func SupportCluster(n string) (yes bool) {
	switch n {
	case KindStructurizr, KindSTAT, KindPlantUML:
		return true
	}

	return false
}
