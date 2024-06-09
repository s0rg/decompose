package graph

import (
	"slices"
	"strconv"

	"github.com/s0rg/decompose/internal/node"
)

type (
	ContainerInfo struct {
		Cmd []string
		Env []string
	}

	VolumeInfo struct {
		Type string
		Src  string
		Dst  string
	}

	Container struct {
		Endpoints map[string]string
		Labels    map[string]string
		conns     map[string]*ConnGroup
		connOrder []string
		ID        string
		Name      string
		Image     string
		Info      *ContainerInfo
		Volumes   []*VolumeInfo
	}
)

func (c *Container) ConnectionsCount() (rv int) {
	for _, cg := range c.conns {
		rv += cg.Len()
	}

	return rv
}

func (c *Container) AddConnection(conn *Connection) {
	if c.conns == nil {
		c.conns = make(map[string]*ConnGroup)
	}

	var seen bool

	grp, seen := c.conns[conn.Process]
	if !seen {
		grp = &ConnGroup{}
	}

	switch {
	case conn.IsListener():
		grp.AddListener(conn)
	case !conn.IsInbound():
		grp.AddOutbound(conn)
	default:
		return
	}

	if !seen {
		c.connOrder = append(c.connOrder, conn.Process)
		slices.Sort(c.connOrder)
	}

	c.conns[conn.Process] = grp
}

func (c *Container) AddMany(conns []*Connection) {
	for _, conn := range conns {
		c.AddConnection(conn)
	}
}

func (c *Container) IterOutbounds(it func(*Connection)) {
	for _, k := range c.connOrder {
		c.conns[k].IterOutbounds(it)
	}
}

func (c *Container) IterListeners(it func(*Connection)) {
	for _, k := range c.connOrder {
		c.conns[k].IterListeners(it)
	}
}

func (c *Container) SortConnections() {
	for _, cg := range c.conns {
		cg.Sort()
	}
}

func (c *Container) ToNode() (rv *node.Node) {
	rv = &node.Node{
		ID:       c.ID,
		Name:     c.Name,
		Image:    c.Image,
		Ports:    &node.Ports{},
		Volumes:  make([]*node.Volume, len(c.Volumes)),
		Networks: make([]string, 0, len(c.Endpoints)),
	}

	for _, n := range c.Endpoints {
		rv.Networks = append(rv.Networks, n)
	}

	c.IterListeners(func(conn *Connection) {
		value := conn.Path

		if conn.Proto != UNIX {
			value = strconv.Itoa(conn.SrcPort)
		}

		rv.Ports.Add(conn.Process, &node.Port{
			Local: conn.IsLocal(),
			Kind:  conn.Proto.String(),
			Value: value,
		})
	})

	for idx, v := range c.Volumes {
		rv.Volumes[idx] = &node.Volume{
			Type: v.Type,
			Src:  v.Src,
			Dst:  v.Dst,
		}
	}

	rv.Container.Labels = c.Labels

	if c.Info != nil {
		rv.Container.Cmd = c.Info.Cmd
		rv.Container.Env = c.Info.Env
	}

	return rv
}
