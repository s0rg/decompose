package node

import (
	"strconv"
	"strings"
)

const portMax = 65535

type Port struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}

func (p *Port) Label() string {
	return strconv.Itoa(p.Value) + "/" + p.Kind
}

func (p *Port) ID() string {
	return p.Kind + strconv.Itoa(p.Value)
}

func (p *Port) Equal(v *Port) (yes bool) {
	return p.Kind == v.Kind && p.Value == v.Value
}

func ParsePort(v string) (rv *Port, ok bool) {
	const (
		nparts = 2
		sep    = "/"
	)

	parts := strings.SplitN(v, sep, nparts)
	if len(parts) != nparts {
		return
	}

	iport, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}

	if iport < 0 || iport > portMax {
		return
	}

	rv = &Port{
		Kind:  parts[1],
		Value: iport,
	}

	return rv, true
}
