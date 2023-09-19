package cluster

import (
	"cmp"
	"slices"
)

type match struct {
	Matcher *Node
	Weight  float64
}

type NodeGrouper struct {
	groups     map[*Node][]string
	matchers   []*Node
	similarity float64
}

func NewGrouper(
	similarity float64,
) *NodeGrouper {
	return &NodeGrouper{
		similarity: similarity,
		matchers:   []*Node{},
		groups:     make(map[*Node][]string),
	}
}

func (ng *NodeGrouper) Add(k string, n *Node) {
	matches := make([]match, 0, len(ng.matchers))

	for _, m := range ng.matchers {
		if w := m.Match(k, n); w >= ng.similarity {
			matches = append(matches, match{
				Matcher: m,
				Weight:  w,
			})
		}
	}

	var (
		best  *Node
		found = true
	)

	switch len(matches) {
	case 0:
		best, found = n.Clone(), false
		ng.matchers = append(ng.matchers, best)
	case 1:
		best = matches[0].Matcher
	default:
		best = slices.MaxFunc(matches, func(a, b match) int {
			return cmp.Compare(a.Weight, b.Weight)
		}).Matcher
	}

	if found {
		best.Merge(n)
	}

	ng.groups[best] = append(ng.groups[best], k)
}

func (ng *NodeGrouper) IterGroups(
	iter func(int, []string),
) {
	for id, m := range ng.matchers {
		if l := ng.groups[m]; len(l) > 0 {
			iter(id, l)
		}
	}
}
