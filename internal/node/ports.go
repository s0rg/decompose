package node

import (
	"cmp"
	"slices"

	"github.com/s0rg/set"
)

type Ports struct {
	ports map[string][]*Port
	order []string
}

func (ps *Ports) Add(process string, p *Port) {
	if ps.ports == nil {
		ps.ports = make(map[string][]*Port)
	}

	ports, ok := ps.ports[process]

	ps.ports[process] = append(ports, p)

	if !ok {
		ps.order = append(ps.order, process)
		slices.Sort(ps.order)
	}
}

func (ps *Ports) Join(other *Ports) {
	other.Iter(func(process string, ports []*Port) {
		for _, p := range ports {
			ps.Add(process, p)
		}
	})
}

func (ps *Ports) Get(p *Port) (name string, ok bool) {
	for name, pl := range ps.ports {
		for _, l := range pl {
			if l.Equal(p) {
				return name, true
			}
		}
	}

	return
}

func (ps *Ports) Iter(it func(process string, p []*Port)) {
	for _, name := range ps.order {
		it(name, ps.ports[name])
	}
}

func (ps *Ports) Len() (rv int) {
	for _, pl := range ps.ports {
		rv += len(pl)
	}

	return rv
}

func (ps *Ports) Sort() {
	for k, pl := range ps.ports {
		slices.SortFunc(pl, func(a, b *Port) int {
			if a.Kind == b.Kind {
				return cmp.Compare(a.Value, b.Value)
			}

			return cmp.Compare(a.Kind, b.Kind)
		})

		ps.ports[k] = slices.CompactFunc(pl, func(a, b *Port) bool {
			return a.Equal(b)
		})
	}
}

/*
	func (ps *Ports) FlatList() (rv []*Port) {
		for _, pl := range ps.ports {
			rv = append(rv, pl...)
		}

		slices.SortFunc(rv, func(a, b *Port) int {
			if a.Kind == b.Kind {
				return cmp.Compare(a.Value, b.Value)
			}

			return cmp.Compare(a.Kind, b.Kind)
		})

		return slices.CompactFunc(rv, func(a, b *Port) bool {
			return a.Equal(b)
		})
	}
*/

func (ps *Ports) HasAny(label ...string) (yes bool) {
	if len(ps.ports) == 0 {
		return
	}

	s := make(set.Unordered[string])
	set.Load(s, label...)

	for _, ports := range ps.ports {
		for _, p := range ports {
			if s.Has(p.Label()) {
				return true
			}
		}
	}

	return false
}

func (ps *Ports) Has(label ...string) (yes bool) {
	if len(ps.ports) == 0 {
		return
	}

	s := make(set.Unordered[string])
	set.Load(s, label...)

	for _, ports := range ps.ports {
		for _, p := range ports {
			s.Del(p.Label())
		}
	}

	return s.Len() == 0
}
