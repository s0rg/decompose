package node

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type portJSON struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
	Local bool   `json:"local"`
}

type Port struct {
	Kind   string `json:"kind"`
	Value  string `json:"value"`
	Number int    `json:"-"`
	Local  bool   `json:"local"`
}

func (p *Port) Label() string {
	return p.Kind + ":" + p.Value
}

func (p *Port) Equal(v *Port) (yes bool) {
	return p.Kind == v.Kind &&
		p.Value == v.Value
}

func (p *Port) UnmarshalJSON(b []byte) (err error) {
	var v portJSON

	if err = json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	if p.Number, err = strconv.Atoi(v.Value); err != nil {
		return fmt.Errorf("invalid port: '%s' atoi: %w", v.Value, err)
	}

	p.Kind = v.Kind
	p.Value = v.Value
	p.Local = v.Local

	return nil
}
