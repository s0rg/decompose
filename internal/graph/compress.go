package graph

import (
	"cmp"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

const externalGroup = "external"

type nodeGroup struct {
	Node  *node.Node
	Group string
}

type Compressor struct {
	b      NamedBuilderWriter
	nodes  map[string]*nodeGroup
	groups map[string][]*node.Node
	conns  map[string]map[string][]*node.Port
	edges  int
}

func NewCompressor(
	bldr NamedBuilderWriter,
) *Compressor {
	return &Compressor{
		b:      bldr,
		nodes:  make(map[string]*nodeGroup),
		groups: make(map[string][]*node.Node),
		conns:  make(map[string]map[string][]*node.Port),
	}
}

func (c *Compressor) AddNode(n *node.Node) error {
	var (
		grp string
		ok  bool
	)

	if n.IsExternal() {
		grp, ok = externalGroup, true
	} else {
		grp, ok = c.find(n.Name)
	}

	if !ok {
		grp = c.extract(n.Name)
	}

	c.groups[grp] = append(c.groups[grp], n)
	c.nodes[n.ID] = &nodeGroup{
		Group: grp,
		Node:  n,
	}

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

	dmap, ok := c.conns[nsrc.Group]
	if !ok {
		dmap = make(map[string][]*node.Port)
		c.conns[nsrc.Group] = dmap
	}

	dmap[ndst.Group] = append(dmap[ndst.Group], e.Port)
	c.edges++
}

func (c *Compressor) Name() string {
	return c.b.Name() + " [compressed]"
}

func (c *Compressor) Write(w io.Writer) (err error) {
	for grp, nodes := range c.groups {
		if err = c.b.AddNode(compressNodes(grp, nodes)); err != nil {
			return fmt.Errorf("compressor add-node [%s]: %w", c.b.Name(), err)
		}
	}

	log.Printf("nodes compressed %d -> %d", len(c.nodes), len(c.groups))

	var count int

	for src, dmap := range c.conns {
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

	log.Printf("edges compressed %d -> %d", c.edges, count)

	if err = c.b.Write(w); err != nil {
		return fmt.Errorf("compressor write [%s]: %w", c.b.Name(), err)
	}

	return nil
}

func (c *Compressor) find(a string) (grp string, ok bool) {
	for grp = range c.groups {
		if c.match(grp, a) {
			return grp, true
		}
	}

	return
}

func (c *Compressor) match(a, b string) bool {
	return strings.HasPrefix(b, a)
}

func (c *Compressor) extract(a string) string {
	const cutset = "0123456789"

	return strings.TrimRight(a, cutset)
}

func compressNodes(id string, nodes []*node.Node) (rv *node.Node) {
	ports := &node.Ports{}
	tags := make([]string, len(nodes))

	for i, n := range nodes {
		tags[i] = n.Name
		ports.Join(n.Ports)
	}

	ports.Sort()

	name := id
	if id != externalGroup {
		name = "group-" + id
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
