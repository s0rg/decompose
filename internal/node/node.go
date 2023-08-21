package node

type Node struct {
	ID       string
	Name     string
	Image    string
	Networks []string
	Meta     *Meta
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
		Connected:  make(map[string][]string),
	}

	if n.Meta != nil {
		rv.Meta = n.Meta
	}

	if n.Image != "" {
		rv.Image = &n.Image
	}

	for i := 0; i < len(n.Ports); i++ {
		rv.Listen[i] = n.Ports[i].Label()
	}

	return rv
}
