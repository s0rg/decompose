package node

type Port struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
	Local bool   `json:"local"`
}

func (p *Port) Label() string {
	return p.Kind + ":" + p.Value
}

func (p *Port) Equal(v *Port) (yes bool) {
	return p.Kind == v.Kind &&
		p.Value == v.Value
}
