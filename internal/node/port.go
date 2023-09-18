package node

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/s0rg/set"
)

type Port struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}

type Ports []*Port

func (p *Port) Label() string {
	return strconv.Itoa(p.Value) + "/" + p.Kind
}

func (p *Port) ID() string {
	return p.Kind + strconv.Itoa(p.Value)
}

func (ps Ports) Dedup() (rv Ports) {
	s := make(set.Unordered[Port])

	for _, p := range ps {
		s.Add(*p)
	}

	p := set.ToSlice(s)
	rv = make([]*Port, len(p))

	for i := 0; i < len(p); i++ {
		rv[i] = &p[i]
	}

	slices.SortStableFunc(rv, func(a, b *Port) int {
		if a.Kind == b.Kind {
			return cmp.Compare(a.Value, b.Value)
		}

		return cmp.Compare(a.Kind, b.Kind)
	})

	return rv
}

func (ps Ports) HasAny(label ...string) (yes bool) {
	if len(ps) == 0 {
		return
	}

	s := make(set.Unordered[string])
	set.Load(s, label...)

	for _, p := range ps {
		if s.Has(p.Label()) {
			return true
		}
	}

	return false
}

func (ps Ports) Has(label ...string) (yes bool) {
	if len(ps) == 0 {
		return
	}

	s := make(set.Unordered[string])
	set.Load(s, label...)

	for _, p := range ps {
		s.Del(p.Label())
	}

	return s.Len() == 0
}
