package cluster

import (
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

type connNode struct {
	Inbounds  set.Unordered[string]
	Outbounds set.Unordered[string]
	Ports     node.Ports
}

type connGraph map[string]*connNode

func (g connGraph) upsert(name string) (gn *connNode) {
	var ok bool

	if gn, ok = g[name]; !ok {
		gn = &connNode{
			Outbounds: make(set.Unordered[string]),
			Inbounds:  make(set.Unordered[string]),
			Ports:     node.Ports{},
		}

		g[name] = gn
	}

	return gn
}

func (g connGraph) AddNode(n *node.Node) {
	gn := g.upsert(n.ID)

	gn.Ports = append(gn.Ports, n.Ports...)
}

func (g connGraph) AddEdge(src, dst string) {
	g.upsert(src).Outbounds.Add(dst)
	g.upsert(dst).Inbounds.Add(src)
}

func (g connGraph) NextLayer(
	from []string,
	seen set.Unordered[string],
) (rv []string) {
	if len(from) == 0 {
		for k, n := range g {
			switch {
			case len(n.Inbounds) > 0:
			case len(n.Ports) == 0:
			default:
				if seen.Add(k) {
					rv = append(rv, k)
				}
			}
		}
	} else {
		set.Load(seen, from...)

		for _, src := range from {
			n := g[src]

			for k := range n.Outbounds {
				if seen.Add(k) {
					rv = append(rv, k)
				}
			}
		}
	}

	slices.Sort(rv)

	return rv
}
