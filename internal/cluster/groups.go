package cluster

import (
	"cmp"
	"slices"

	"github.com/s0rg/decompose/internal/node"
)

type match struct {
	Matcher *PortMatcher
	Weight  float64
	Index   int
}

type PortGrouper struct {
	groups     map[*PortMatcher][]string
	matchers   []*PortMatcher
	similarity float64
}

func NewGrouper(
	similarity float64,
) *PortGrouper {
	return &PortGrouper{
		similarity: similarity,
		matchers:   []*PortMatcher{},
		groups:     make(map[*PortMatcher][]string),
	}
}

func (pg *PortGrouper) Add(
	key string,
	ports node.Ports,
) {
	current := FromPorts(ports)
	matches := make([]match, 0, len(pg.matchers))

	for _, m := range pg.matchers {
		if w := m.Match(current); w >= pg.similarity {
			matches = append(matches, match{
				Matcher: m,
				Weight:  w,
			})
		}
	}

	var best *PortMatcher

	switch len(matches) {
	case 0:
		pg.matchers = append(pg.matchers, current)
		best = current
	case 1:
		best = matches[0].Matcher
	default:
		best = slices.MaxFunc(matches, func(a, b match) int {
			return cmp.Compare(a.Weight, b.Weight)
		}).Matcher
	}

	if best != current {
		best.Merge(current)
	}

	pg.groups[best] = append(pg.groups[best], key)
}

func (pg *PortGrouper) IterGroups(
	iter func(id int, members []string),
) {
	for id, m := range pg.matchers {
		if l := pg.groups[m]; len(l) > 0 {
			iter(id, l)
		}
	}
}
