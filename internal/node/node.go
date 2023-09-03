package node

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
		rv.Meta = n.Meta
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
