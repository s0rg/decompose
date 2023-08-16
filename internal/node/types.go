package node

import "strconv"

type Port struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}

func (p *Port) Label() string {
	return strconv.Itoa(p.Value) + "/" + p.Kind
}

func (p *Port) String() string {
	return strconv.Itoa(p.Value)
}

type Node struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	Ports []Port `json:"ports"`
}

func (n *Node) IsExternal() bool {
	return n.ID == n.Name
}
