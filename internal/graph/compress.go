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
)

type Compressor struct {
	b      NamedBuilderWriter
	nodes  map[string]*node.Node              // "raw" incoming nodes nodeID -> node
	groups map[string]*node.Node              // "compressed" nodes groupID -> node
	index  map[string]string                  // index holds nodeID -> groupID mapping
	conns  map[string]map[string][]*node.Port // "raw" connections nodeID -> nodeID -> []port
	edges  int
	diff   int
	force  bool
}

func NewCompressor(
	bldr NamedBuilderWriter,
	diff int,
	force bool,
) *Compressor {
	return &Compressor{
		b:      bldr,
		diff:   diff,
		force:  force,
		index:  make(map[string]string),
		nodes:  make(map[string]*node.Node),
		groups: make(map[string]*node.Node),
		conns:  make(map[string]map[string][]*node.Port),
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
	c.buildGroups()
	edges, count := c.buildEdges()

	for _, node := range c.groups {
		if err = c.b.AddNode(node); err != nil {
			return fmt.Errorf("compressor add-node [%s]: %w", c.b.Name(), err)
		}
	}

	log.Printf("[compress] nodes %d -> %d %.02f%%",
		len(c.nodes),
		len(c.groups),
		percentOf(len(c.nodes)-len(c.groups), len(c.nodes)),
	)

	for src, dmap := range edges {
		for dst, ports := range dmap {
			for _, port := range ports {
				c.b.AddEdge(&node.Edge{
					SrcID: src,
					DstID: dst,
					Port:  port,
				})
			}
		}
	}

	log.Printf("[compress] edges %d -> %d %.02f%%",
		c.edges,
		count,
		percentOf(c.edges-count, c.edges),
	)

	if err = c.b.Write(w); err != nil {
		return fmt.Errorf("compressor write [%s]: %w", c.b.Name(), err)
	}

	return nil
}

func (c *Compressor) buildGroups() {
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

	comm := t.Common("", c.diff)

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
			c.index[nodeID] = grp

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
		c.index[id] = grp

		return true
	})

	for grp, nodes := range groups {
		batch := make([]*node.Node, len(nodes))

		for i, nodeID := range nodes {
			batch[i] = c.nodes[nodeID]
		}

		c.groups[grp] = compressNodes(grp, batch)
	}
}

func (c *Compressor) buildEdges() (
	edges map[string]map[string][]*node.Port,
	count int,
) {
	edges = make(map[string]map[string][]*node.Port)

	// initial compression: compress to groups
	for src, dmap := range c.conns {
		srcg := c.index[src]

		gmap, ok := edges[srcg]
		if !ok {
			gmap = make(map[string][]*node.Port)
			edges[srcg] = gmap
		}

		for dst, ports := range dmap {
			if src == dst { // skip nodes cycles
				continue
			}

			dstg := c.index[dst]
			if srcg == dstg {
				continue // skip groups cycles
			}

			gmap[dstg] = append(gmap[dstg], ports...)
		}
	}

	if c.force {
		edges = c.forceCompress(edges)
	}

	for _, dmap := range edges {
		for dst, ports := range dmap {
			ports = compressPorts(ports)
			dmap[dst] = ports
			count += len(ports)
		}
	}

	return edges, count
}

// force compression: remove single-connected groups.
func (c *Compressor) forceCompress(
	edges map[string]map[string][]*node.Port,
) (
	rv map[string]map[string][]*node.Port,
) {
	dsts := make(map[string][]string)
	drop := func(k, v string) {
		delete(c.groups, v)
		delete(edges[k], v)
		delete(edges, v)
	}

	for src, dmap := range edges {
		if len(dmap) == 1 {
			for key := range dmap {
				drop(key, src)
			}

			continue
		}

		for dst := range dmap {
			dsts[dst] = append(dsts[dst], src)
		}
	}

	for k, v := range dsts {
		if len(v) == 1 {
			drop(v[0], k)
		}
	}

	return edges
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
