package cluster

import (
	"math"
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

type Node struct {
	Inbounds  set.Set[string]
	Outbounds set.Set[string]
	Ports     *node.Ports
}

func (n *Node) Clone() *Node {
	return &Node{
		Inbounds:  n.Inbounds.Clone(),
		Outbounds: n.Outbounds.Clone(),
		Ports:     n.Ports,
	}
}

const (
	onei = 1
	onef = 1.0
	half = 0.5
)

func (n *Node) Match(id string, o *Node) (rv float64) {
	rv = (n.matchConns(id) + n.matchPorts(o.Ports)) * half

	return rv
}

func (n *Node) Merge(o *Node) {
	n.Ports.Join(o.Ports)
	n.Ports.Sort()

	n.Inbounds = set.Union(n.Inbounds, o.Inbounds)
	n.Outbounds = set.Union(n.Outbounds, o.Outbounds)
}

func (n *Node) matchConns(id string) (rv float64) {
	if n.Inbounds.Has(id) {
		rv += half
	}

	if n.Outbounds.Has(id) {
		rv += half
	}

	return rv
}

func (n *Node) matchPorts(p *node.Ports) (rv float64) {
	var (
		a = portsToProtos(n.Ports)
		b = portsToProtos(p)
	)

	if len(a) > len(b) {
		a, b = b, a
	}

	for k, ap := range a {
		bp := b[k]

		rv += matchSlices(ap, bp) / float64(len(a))
	}

	return rv
}

func portsToProtos(ports *node.Ports) (rv map[string][]int) {
	rv = make(map[string][]int)

	ports.Iter(func(_ string, pl []*node.Port) {
		for _, p := range pl {
			rv[p.Kind] = append(rv[p.Kind], p.Value)
		}
	})

	for k := range rv {
		slices.Sort(rv[k])
	}

	return rv
}

func matchSlices(a, b []int) (rv float64) {
	sa, sb := make(set.Unordered[int]), make(set.Unordered[int])

	if len(a) < len(b) {
		a, b = b, a
	}

	set.Load(sa, a...)
	set.Load(sb, b...)

	u := set.Union(sa, sb).Len()
	c := set.Intersect(sa, sb).Len()

	if u == c {
		return onef
	}

	da, db := set.Diff(sa, sb), set.Diff(sb, sa)

	if da.Len() < db.Len() {
		da, db = db, da
	}

	rv = float64(da.Len()) / math.Abs(float64(u)-float64(c))

	das, dab := set.ToSlice(da), set.ToSlice(db)

	slices.Sort(das)
	slices.Sort(dab)

	m := float64(db.Len()) / float64(u)

	for i := 0; i < len(das); i++ {
		for j := 0; j < len(dab); j++ {
			if abs(das[i]-dab[j]) == onei {
				rv += m
			}
		}
	}

	return rv
}

func abs(v int) int {
	if v < 0 {
		v = -v
	}

	return v
}
