package node

import (
	"cmp"
	"slices"
)

type Ports struct {
	ports map[string][]*Port
}

func (ps *Ports) Add(process string, p *Port) {
	if ps.ports == nil {
		ps.ports = make(map[string][]*Port)
	}

	ps.ports[process] = append(ps.ports[process], p)
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
	for name, pl := range ps.ports {
		it(name, pl)
	}

	return
}

func (ps *Ports) Len() (rv int) {
	for _, pl := range ps.ports {
		rv += len(pl)
	}

	return rv
}

func (ps *Ports) Sort() {
	for _, pl := range ps.ports {
		slices.SortFunc(pl, func(a, b *Port) int {
			if a.Kind == b.Kind {
				return cmp.Compare(a.Value, b.Value)
			}

			return cmp.Compare(a.Kind, b.Kind)
		})
	}

	return
}

func (ps *Ports) HasAny(label ...string) (yes bool) {
	/*
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
	*/
	return
}

func (ps *Ports) Has(label ...string) (yes bool) {
	/*
		    if len(ps) == 0 {
				return
			}

			s := make(set.Unordered[string])
			set.Load(s, label...)

			for _, p := range ps {
				s.Del(p.Label())
			}

			return s.Len() == 0
	*/
	return
}
