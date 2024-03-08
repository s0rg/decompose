package node

import (
	"path/filepath"
	"slices"
)

type Node struct {
	Container Container
	Meta      *Meta
	Ports     *Ports
	ID        string
	Name      string
	Image     string
	Cluster   string
	Networks  []string
	Volumes   []*Volume
}

func External(name string) (rv *Node) {
	return &Node{
		ID:    name,
		Name:  name,
		Ports: &Ports{},
	}
}

func (n *Node) IsExternal() bool {
	return n.ID == n.Name
}

func (n *Node) ToJSON() (rv *JSON) {
	rv = &JSON{
		Name:       n.Name,
		IsExternal: n.IsExternal(),
		Networks:   n.Networks,
		Container:  n.Container,
		Listen:     make(map[string][]*Port),
		Volumes:    []*Volume{},
		Tags:       []string{},
		Connected:  make(map[string][]*Connection),
	}

	if n.Meta != nil {
		rv.Tags = n.Meta.Tags
	}

	if n.Image != "" {
		rv.Image = &n.Image
	}

	n.Ports.Sort()

	n.Ports.Iter(func(name string, ports []*Port) {
		rv.Listen[name] = ports
	})

	if len(n.Volumes) > 0 {
		rv.Volumes = n.Volumes
	}

	return rv
}

func (n *Node) ToView() (rv *View) {
	rv = &View{
		Name:       n.Name,
		Image:      n.Image,
		Listen:     n.Ports,
		IsExternal: n.IsExternal(),
	}

	if n.Meta != nil && len(n.Meta.Tags) > 0 {
		rv.Tags = n.Meta.Tags
	}

	if len(n.Container.Cmd) > 0 {
		rv.Cmd = filepath.Base(n.Container.Cmd[0])
		rv.Args = n.Container.Cmd[1:]
	}

	return rv
}

func (n *Node) FormatMeta() (rv []string, ok bool) {
	if n.Meta == nil {
		return
	}

	const maxMeta = 3

	rv = make([]string, 0, maxMeta)

	if n.Meta.Info != "" {
		rv = append(rv, n.Meta.Info)
	}

	if n.Meta.Docs != "" {
		rv = append(rv, n.Meta.Docs)
	}

	if n.Meta.Repo != "" {
		rv = append(rv, n.Meta.Repo)
	}

	return slices.Clip(rv), true
}
