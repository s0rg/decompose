package cluster

import (
	"slices"

	"github.com/s0rg/decompose/internal/node"
	"github.com/s0rg/set"
)

type PortMatcher struct {
	ports node.Ports
}

func FromPorts(
	ports node.Ports,
) *PortMatcher {
	return &PortMatcher{
		ports: slices.Clone(ports),
	}
}

func portsToProtos(ports node.Ports) (rv map[string][]int) {
	rv = make(map[string][]int)

	for _, p := range ports {
		rv[p.Kind] = append(rv[p.Kind], p.Value)
	}

	for k := range rv {
		slices.Sort(rv[k])
	}

	return rv
}

func (pm *PortMatcher) Match(other *PortMatcher) (rv float64) {
	var (
		a = portsToProtos(pm.ports)
		b = portsToProtos(other.ports)
	)

	if len(a) > len(b) {
		a, b = b, a
	}

	const (
		one = 1
		two = 2.0
	)

	for k, ap := range a {
		bp := b[k]

		sa := make(set.Unordered[int])
		sb := make(set.Unordered[int])

		if len(ap) >= len(bp) {
			ap, bp = bp, ap
		}

		set.Load(sa, ap...)
		set.Load(sb, bp...)

		m := set.Intersect(sa, sb).Len()
		u := float64(set.Union(sa, sb).Len()) / two

		for _, av := range ap {
			for _, bv := range bp {
				if abs(av-bv) == one {
					m++
				}
			}
		}

		rv += float64(m) / u
	}

	return rv
}

func (pm *PortMatcher) Merge(other *PortMatcher) {
	pm.ports = append(pm.ports, other.ports...).Dedup()
}

func abs(v int) int {
	if v < 0 {
		v = -v
	}

	return v
}
