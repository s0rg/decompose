package cluster

import (
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

type Node struct {
	Inbounds  set.Set[string]
	Outbounds set.Set[string]
	Ports     node.Ports
}

func (n *Node) Clone() *Node {
	return &Node{
		Inbounds:  n.Inbounds.Clone(),
		Outbounds: n.Outbounds.Clone(),
		Ports:     slices.Clone(n.Ports),
	}
}

func (n *Node) Match(id string, o *Node) (rv float64) {
	rv = n.matchConns(id) + n.matchPorts(o.Ports)

	return rv
}

func (n *Node) Merge(o *Node) {
	n.Ports = append(n.Ports, o.Ports...).Dedup()
	n.Inbounds = set.Union(n.Inbounds, o.Inbounds)
	n.Outbounds = set.Union(n.Outbounds, o.Outbounds)
}

func (n *Node) matchConns(id string) (rv float64) {
	t := n.Inbounds.Len() + n.Outbounds.Len()
	if t == 0 {
		return
	}

	if n.Inbounds.Has(id) {
		rv += 1.0 / float64(n.Inbounds.Len())
	}

	if n.Outbounds.Has(id) {
		rv += 1.0 / float64(n.Outbounds.Len())
	}

	return rv / float64(t)
}

func (n *Node) matchPorts(p node.Ports) (rv float64) {
	var (
		a = portsToProtos(n.Ports)
		b = portsToProtos(p)
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

func abs(v int) int {
	if v < 0 {
		v = -v
	}

	return v
}
