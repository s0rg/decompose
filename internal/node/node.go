package node

import (
	"path/filepath"
	"slices"
)

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
	Info string   `json:"info"`
	Docs string   `json:"docs"`
	Repo string   `json:"repo"`
	Tags []string `json:"tags"`
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
		Tags:       []string{},
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
		Name:       n.Name,
		Image:      n.Image,
		Listen:     n.Ports,
		IsExternal: n.IsExternal(),
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
