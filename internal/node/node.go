package node

import "path/filepath"

type Process struct {
	Cmd []string `json:"cmd"`
	Env []string `json:"env"`
}

type Volume struct {
	Type string `json:"type"`
	Src  string `json:"src"`
	Dst  string `json:"dst"`
}

type Meta struct {
	Info string
	Tags []string
}

type Node struct {
	ID       string
	Name     string
	Image    string
	Cluster  string
	Networks []string
	Meta     *Meta
	Process  *Process
	Volumes  []*Volume
	Ports    Ports
}

func (n *Node) IsExternal() bool {
	return n.ID == n.Name
}

func (n *Node) ToJSON() (rv *JSON) {
	rv = &JSON{
		Name:       n.Name,
		IsExternal: n.IsExternal(),
		Networks:   n.Networks,
		Listen:     make([]string, len(n.Ports)),
		Volumes:    []*Volume{},
		Connected:  make(map[string][]string),
	}

	if n.Meta != nil {
		rv.Tags = n.Meta.Tags
	}

	if n.Process != nil {
		rv.Process = n.Process
	}

	if n.Image != "" {
		rv.Image = &n.Image
	}

	for i := 0; i < len(n.Ports); i++ {
		rv.Listen[i] = n.Ports[i].Label()
	}

	if len(n.Volumes) > 0 {
		rv.Volumes = n.Volumes
	}

	return rv
}

func (n *Node) ToView() (rv *View) {
	rv = &View{
		Name:  n.Name,
		Image: n.Image,
		Ports: n.Ports,
		Local: !n.IsExternal(),
	}

	if n.Meta != nil && len(n.Meta.Tags) > 0 {
		rv.Tags = n.Meta.Tags
	}

	if n.Process != nil && len(n.Process.Cmd) > 0 {
		rv.Cmd = filepath.Base(n.Process.Cmd[0])
		rv.Args = n.Process.Cmd[1:]
	}

	return rv
}
