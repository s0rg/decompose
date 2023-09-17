package cluster

import (
	"cmp"
	"slices"

	"github.com/s0rg/decompose/internal/node"
	"github.com/s0rg/set"
)

type match struct {
	Matcher *PortMatcher
	Weight  float64
	Index   int
}

func (m *match) Match(o match) (rv float64) {
	return m.Matcher.Match(o.Matcher)
}

func (m *match) Merge(o match) {
	m.Matcher.Merge(o.Matcher)
}

type PortGrouper struct {
	groups     map[*PortMatcher][]string
	matchers   []*PortMatcher
	confidence float64
}

func NewGrouper(
	confidence float64,
) *PortGrouper {
	return &PortGrouper{
		confidence: confidence,
		matchers:   []*PortMatcher{},
		groups:     make(map[*PortMatcher][]string),
	}
}

func (pg *PortGrouper) Add(
	key string,
	ports set.Unordered[node.Port],
) {
	current := FromPorts(ports)
	matches := make([]match, 0, len(pg.matchers))

	for _, m := range pg.matchers {
		if w := m.Match(current); w >= pg.confidence {
			matches = append(matches, match{
				Matcher: m,
				Weight:  w,
			})
		}
	}

	switch len(matches) {
	case 0:
		pg.matchers = append(pg.matchers, current)
	case 1:
		current = matches[0].Matcher
	default:
		current = slices.MaxFunc(matches, func(a, b match) int {
			return cmp.Compare(a.Weight, b.Weight)
		}).Matcher
	}

	pg.groups[current] = append(pg.groups[current], key)
}

/*
	func (pg *PortGrouper) String() (rv string) {
		var b bytes.Buffer

		for id, m := range pg.matchers {
			group := pg.groups[m]
			if len(group) == 0 {
				continue
			}

			fmt.Fprintf(&b, "matcher[%d] ports: %s\n", id, m)

			for _, name := range group {
				fmt.Fprintln(&b, "\t", name)
			}

			fmt.Fprintln(&b, "")
		}

		return b.String()
	}
*/
func (pg *PortGrouper) Groups() int {
	return len(pg.groups)
}

func (pg *PortGrouper) Count() (rv int) {
	for _, g := range pg.groups {
		rv += len(g)
	}

	return rv
}

func (pg *PortGrouper) sortedMatches() (rv []match) {
	rv = make([]match, 0, len(pg.matchers))

	for _, m := range pg.matchers {
		rv = append(rv, match{Matcher: m})
	}

	rv = slices.Clip(rv)

	slices.SortStableFunc(rv, func(a, b match) int {
		return cmp.Compare(b.Matcher.Count(), a.Matcher.Count())
	})

	return rv
}

func (pg *PortGrouper) singleGroups() (rv []match) {
	rv = make([]match, 0, len(pg.groups))

	for k, g := range pg.groups {
		if len(g) > 1 || k.Count() > 1 {
			continue
		}

		rv = append(rv, match{Matcher: k})
	}

	return slices.Clip(rv)
}

func (pg *PortGrouper) merge(a, b *PortMatcher) {
	pg.groups[a] = append(pg.groups[a], pg.groups[b]...)
	delete(pg.groups, b)
	a.Merge(b)
}

func (pg *PortGrouper) compress(
	matches []match,
	domatch func(a, b *PortMatcher) bool,
) {
	for changed := true; changed; {
		changed = false

		for i := 0; i < len(matches); i++ {
			curr := matches[i]

			found := make([]match, 0, len(matches)-i)

			for j := i + 1; j < len(matches); j++ {
				if c := matches[j]; domatch(curr.Matcher, c.Matcher) {
					c.Index, c.Weight = j, curr.Match(c)
					found = append(found, c)
				}
			}

			var best match

			switch len(found) {
			case 0:
				continue
			case 1:
				best = found[0]
			default:
				best = slices.MaxFunc(found, func(a, b match) int {
					return cmp.Compare(a.Weight, b.Weight)
				})
			}

			pg.merge(curr.Matcher, best.Matcher)

			matches = slices.Delete(matches, best.Index, best.Index+1)
			changed = true

			break
		}
	}
}

func (pg *PortGrouper) Compress() {
	// step 1: compress by inclusion
	pg.compress(pg.sortedMatches(), func(a, b *PortMatcher) bool {
		return a.Contains(b) >= pg.confidence
	})

	// step 2: compress by closest port range
	pg.compress(pg.sortedMatches(), func(a, b *PortMatcher) bool {
		return a.Diff(b) >= 0.5
	})
}
