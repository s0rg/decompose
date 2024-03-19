package graph

import (
	"cmp"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"

	"github.com/s0rg/set"
	"github.com/s0rg/trie"

	"github.com/s0rg/decompose/internal/node"
)

const (
	externalGroup = "external"
	defaultGroup  = "core"
	defaultDiff   = 3
)

type Compressor struct {
	b     NamedBuilderWriter
	nodes map[string]*node.Node
	conns map[string]map[string][]*node.Port
	edges int
}

func NewCompressor(
	bldr NamedBuilderWriter,
) *Compressor {
	return &Compressor{
		b:     bldr,
		nodes: make(map[string]*node.Node),
		conns: make(map[string]map[string][]*node.Port),
	}
}

func (c *Compressor) AddNode(n *node.Node) error {
	c.nodes[n.ID] = n

	return nil
}

func (c *Compressor) AddEdge(e *node.Edge) {
	nsrc, ok := c.nodes[e.SrcID]
	if !ok {
		return
	}

	ndst, ok := c.nodes[e.DstID]
	if !ok {
		return
	}

	dmap, ok := c.conns[nsrc.ID]
	if !ok {
		dmap = make(map[string][]*node.Port)
		c.conns[nsrc.ID] = dmap
	}

	dmap[ndst.ID] = append(dmap[ndst.ID], e.Port)
	c.edges++
}

func (c *Compressor) Name() string {
	return c.b.Name() + " [compressed]"
}

func (c *Compressor) Write(w io.Writer) (err error) {
	index, err := c.buildGroups()
	if err != nil {
		return err
	}

	c.buildEdges(index)

	if err = c.b.Write(w); err != nil {
		return fmt.Errorf("compressor write [%s]: %w", c.b.Name(), err)
	}

	return nil
}

func (c *Compressor) buildGroups() (index map[string]string, err error) {
	index = make(map[string]string)
	groups := make(map[string][]string)

	t := trie.New[string]()
	seen := make(set.Unordered[string])

	for _, node := range c.nodes {
		seen.Add(node.ID)

		if node.IsExternal() {
			continue
		}

		t.Add(node.Name, node.ID)
	}

	comm := t.Common("", defaultDiff)

	for _, key := range comm {
		nodes := []string{}

		t.Iter(key, func(_, nodeID string) {
			nodes = append(nodes, nodeID)
		})

		grp := defaultGroup
		if len(nodes) > 1 {
			grp = cleanName(key)
		}

		for _, nodeID := range nodes {
			index[nodeID] = grp

			seen.Del(nodeID)
		}

		groups[grp] = nodes
	}

	seen.Iter(func(id string) (next bool) {
		grp := defaultGroup

		if c.nodes[id].IsExternal() {
			grp = externalGroup
		}

		groups[grp] = append(groups[grp], id)
		index[id] = grp

		return true
	})

	for grp, nodes := range groups {
		batch := make([]*node.Node, len(nodes))

		for i, nodeID := range nodes {
			batch[i] = c.nodes[nodeID]
		}

		if err = c.b.AddNode(compressNodes(grp, batch)); err != nil {
			err = fmt.Errorf("compressor add-node [%s]: %w", c.b.Name(), err)

			return nil, err
		}
	}

	log.Printf("[compress] nodes %d -> %d %.02f%%",
		len(c.nodes),
		len(groups),
		percentOf(len(c.nodes)-len(groups), len(c.nodes)),
	)

	return index, nil
}

func (c *Compressor) buildEdges(index map[string]string) {
	gconns := make(map[string]map[string][]*node.Port)

	for src, dmap := range c.conns {
		srcg := index[src]

		gmap, ok := gconns[srcg]
		if !ok {
			gmap = make(map[string][]*node.Port)
			gconns[srcg] = gmap
		}

		for dst, ports := range dmap {
			dstg := index[dst]

			gmap[dstg] = append(gmap[dstg], ports...)
		}
	}

	var count int

	for src, dmap := range gconns {
		for dst, ports := range dmap {
			ports = compressPorts(ports)

			for _, port := range ports {
				c.b.AddEdge(&node.Edge{
					SrcID: src,
					DstID: dst,
					Port:  port,
				})

				count++
			}
		}
	}

	log.Printf("[compress] edges %d -> %d %.02f%%",
		c.edges,
		count,
		percentOf(c.edges-count, c.edges),
	)
}

func cleanName(a string) string {
	const cutset = "0123456789-"

	return strings.TrimRight(a, cutset)
}

func compressNodes(id string, nodes []*node.Node) (rv *node.Node) {
	ports := &node.Ports{}
	tags := make([]string, len(nodes))

	for i, n := range nodes {
		tags[i] = n.Name

		n.Ports.Iter(func(_ string, plist []*node.Port) {
			for _, p := range plist {
				ports.Add(n.Name, p)
			}
		})
	}

	ports.Compact()

	name := id
	if name != externalGroup {
		name = strings.ToUpper(name)
	}

	return &node.Node{
		ID:    id,
		Name:  name,
		Ports: ports,
		Meta: &node.Meta{
			Tags: tags,
		},
	}
}

func compressPorts(ports []*node.Port) (rv []*node.Port) {
	slices.SortFunc(ports, func(a, b *node.Port) int {
		if a.Kind == b.Kind {
			return cmp.Compare(a.Value, b.Value)
		}

		return cmp.Compare(a.Kind, b.Kind)
	})

	return slices.CompactFunc(ports, func(a, b *node.Port) bool {
		return a.Equal(b)
	})
}
