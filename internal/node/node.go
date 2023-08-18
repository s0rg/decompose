package node

type Node struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Image    string   `json:"image"`
	Ports    Ports    `json:"ports"`
	Networks []string `json:"networks"`
}

func (n *Node) IsExternal() bool {
	return n.ID == n.Name
}

func (n *Node) ToJSON() (rv *JSON) {
	rv = &JSON{
		Name:       n.Name,
		IsExternal: n.IsExternal(),
		Networks:   make([]string, len(n.Networks)),
		Listen:     make([]string, len(n.Ports)),
		Connected:  make(map[string][]string),
	}

	if n.Image != "" {
		rv.Image = &n.Image
	}

	for i := 0; i < len(n.Ports); i++ {
		rv.Listen[i] = n.Ports[i].Label()
	}

	copy(rv.Networks, n.Networks)

	return rv
}
