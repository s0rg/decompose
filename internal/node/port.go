package node

import (
	"strconv"
)

type Port struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
	Local bool   `json:"local"`
}

func (p *Port) Label() string {
	return strconv.Itoa(p.Value) + "/" + p.Kind
}

func (p *Port) ID() string {
	return p.Kind + strconv.Itoa(p.Value)
}

func (p *Port) Equal(v *Port) (yes bool) {
	return p.Kind == v.Kind &&
		p.Value == v.Value
}
