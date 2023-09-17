package cluster

import (
	"slices"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/node"
)

type connNode struct {
	Ports     set.Unordered[node.Port]
	Inbounds  set.Unordered[string]
	Outbounds set.Unordered[string]
}

type connGraph map[string]*connNode

func (g connGraph) upsert(name string) (gn *connNode) {
	var ok bool

	if gn, ok = g[name]; !ok {
		gn = &connNode{
			Outbounds: make(set.Unordered[string]),
			Inbounds:  make(set.Unordered[string]),
			Ports:     make(set.Unordered[node.Port]),
		}

		g[name] = gn
	}

	return gn
}

func (g connGraph) AddNode(n *node.Node) {
	gn := g.upsert(n.ID)

	for _, p := range n.Ports {
		gn.Ports.Add(*p)
	}
}

func (g connGraph) AddEdge(src, dst string) {
	g.upsert(src).Outbounds.Add(dst)
	g.upsert(dst).Inbounds.Add(src)
}

func (g connGraph) DelNode(name string) {
	delete(g, name)

	for _, node := range g {
		node.Inbounds.Del(name)
		node.Outbounds.Del(name)
	}
}

func (g connGraph) NextLayer(from []string) (rv []string) {
	seen := make(set.Unordered[string])

	if len(from) == 0 {
		for k, node := range g {
			switch {
			case len(node.Inbounds) > 0:
			case len(node.Ports) == 0:
			default:
				if seen.Add(k) {
					rv = append(rv, k)
				}
			}
		}
	} else {
		set.Load(seen, from...)

		for _, src := range from {
			node, ok := g[src]
			if !ok {
				continue
			}

			for k := range node.Outbounds {
				if seen.Add(k) {
					rv = append(rv, k)
				}
			}
		}
	}

	slices.Sort(rv)

	return rv
}
