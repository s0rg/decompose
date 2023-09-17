package cluster

import (
	"cmp"
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

type PortMatcher struct {
	ports set.Set[node.Port]
}

func FromPorts(
	ports set.Unordered[node.Port],
) *PortMatcher {
	return &PortMatcher{
		ports: ports.Clone(),
	}
}

func (pm *PortMatcher) Match(other *PortMatcher) (rv float64) {
	match := set.Intersect(pm.ports, other.ports).
		Len()
	union := set.Union(pm.ports, other.ports).
		Len()

	return float64(match) / float64(union)
}

func (pm *PortMatcher) Contains(other *PortMatcher) (rv float64) {
	match := set.Intersect(pm.ports, other.ports).
		Len()

	return float64(match) / float64(other.Count())
}

func (pm *PortMatcher) Diff(other *PortMatcher) (rv float64) {
	match := set.Intersect(pm.ports, other.ports)
	union := set.Union(pm.ports, other.ports).
		Len() - match.Len()

	a := set.Diff(pm.ports, match)
	b := set.Diff(other.ports, match)
	c := 0

	sa, sb := set.ToSlice(a), set.ToSlice(b)

	if len(sb) > len(sa) {
		sa, sb = sb, sa
	}

	byProto := func(a, b node.Port) int {
		switch {
		case a.Kind == b.Kind:
			return cmp.Compare(a.Value, b.Value)
		case a.Kind == "tcp":
			return 1
		default:
			return -1
		}
	}

	slices.SortStableFunc(sa, byProto)
	slices.SortStableFunc(sb, byProto)

	for i := 0; i < len(sa); i++ {
		for j := 0; j < len(sb); j++ {
			if abs(sa[i].Value-sb[j].Value) == 1 {
				c++
			}
		}
	}

	return float64(c) / float64(union)
}

func (pm *PortMatcher) Merge(other *PortMatcher) {
	other.ports.Iter(func(v node.Port) bool {
		pm.ports.Add(v)

		return true
	})
}

func (pm *PortMatcher) Count() (rv int) {
	return pm.ports.Len()
}

func abs(v int) int {
	if v < 0 {
		v = -v
	}

	return v
}
