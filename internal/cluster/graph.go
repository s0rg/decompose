package cluster

import (
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

const clusterPorts = "[cluster]"

type connGraph map[string]*Node

func (g connGraph) AddNode(n *node.Node) {
	gn := g.upsert(n.ID)

	gn.Ports.Join(n.Ports)
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
			case n.Inbounds.Len() > 0:
			case n.Ports.Len() == 0:
			default:
				if seen.Add(k) {
					rv = append(rv, k)
				}
			}
		}
	} else {
		set.Load(seen, from...)

		for _, src := range from {
			g[src].Outbounds.Iter(func(v string) bool {
				if seen.Add(v) {
					rv = append(rv, v)
				}

				return true
			})
		}
	}

	slices.Sort(rv)

	return rv
}

func (g connGraph) upsert(id string) (gn *Node) {
	var ok bool

	if gn, ok = g[id]; !ok {
		gn = &Node{
			Outbounds: make(set.Unordered[string]),
			Inbounds:  make(set.Unordered[string]),
			Ports:     &node.Ports{},
		}

		g[id] = gn
	}

	return gn
}
