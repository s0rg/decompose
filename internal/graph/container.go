package graph

import (
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
		conns     map[string]*connGroup
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
		c.conns = make(map[string]*connGroup)
	}

	grp, ok := c.conns[conn.Process]
	if !ok {
		grp = &connGroup{}
		c.conns[conn.Process] = grp
	}

	switch {
	case conn.IsListener():
		grp.AddListener(conn)
	case !conn.IsInbound():
		grp.AddOutbound(conn)
	}
}

func (c *Container) AddMany(conns []*Connection) {
	for _, conn := range conns {
		c.AddConnection(conn)
	}
}

func (c *Container) IterOutbounds(it func(*Connection)) {
	for _, cg := range c.conns {
		cg.IterOutbounds(it)
	}
}

func (c *Container) IterListeners(it func(*Connection)) {
	for _, cg := range c.conns {
		cg.IterListeners(it)
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
		Volumes:  []*node.Volume{},
		Networks: make([]string, 0, len(c.Endpoints)),
	}

	for _, n := range c.Endpoints {
		rv.Networks = append(rv.Networks, n)
	}

	c.IterListeners(func(conn *Connection) {
		rv.Ports.Add(conn.Process, &node.Port{
			Kind:  conn.Proto.String(),
			Value: int(conn.LocalPort),
		})
	})

	if len(c.Volumes) > 0 {
		rv.Volumes = make([]*node.Volume, len(c.Volumes))

		for idx, v := range c.Volumes {
			rv.Volumes[idx] = &node.Volume{
				Type: v.Type,
				Src:  v.Src,
				Dst:  v.Dst,
			}
		}
	}

	rv.Container.Labels = c.Labels

	if c.Info != nil {
		rv.Container.Cmd = c.Info.Cmd
		rv.Container.Env = c.Info.Env
	}

	return rv
}
