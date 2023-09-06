package node

import (
	"slices"
	"sort"
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
	state := make(map[string]set.Set[int])

	for i := 0; i < len(ps); i++ {
		p := ps[i]

		s, ok := state[p.Kind]
		if !ok {
			s = make(set.Unordered[int])

			state[p.Kind] = s
		}

		s.Add(p.Value)
	}

	var (
		total int
		keys  = make([]string, 0, len(state))
	)

	for k, s := range state {
		keys = append(keys, k)

		total += s.Len()
	}

	rv = make([]*Port, 0, total)

	slices.Sort(keys)

	for _, k := range keys {
		s := state[k]

		ports := set.ToSlice(s)

		if len(ports) > 1 {
			sort.Ints(ports)
		}

		for _, p := range ports {
			rv = append(rv, &Port{
				Kind:  k,
				Value: p,
			})
		}
	}

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
