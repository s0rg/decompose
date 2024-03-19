package srtructurizr

import (
	"io"
	"slices"
)

type System struct {
	containers    map[string]*Container
	relationships map[string]map[string]*Relation
	ID            string
	Name          string
	Description   string
	Tags          []string
	order         []string
}

func NewSystem(name string) *System {
	return &System{
		ID:            SafeID(name),
		Name:          name,
		containers:    make(map[string]*Container),
		relationships: make(map[string]map[string]*Relation),
	}
}

func (s *System) AddContainer(id, name string) (c *Container, ok bool) {
	id = SafeID(id)

	if _, ok = s.containers[id]; ok {
		return nil, false
	}

	c = &Container{
		ID:   id,
		Name: name,
	}

	s.containers[id] = c
	s.order = append(s.order, id)
	slices.Sort(s.order)

	return c, true
}

func (s *System) findRelation(src, dst string) (rv *Relation, found bool) {
	src, dst = SafeID(src), SafeID(dst)

	if dest, ok := s.relationships[src]; ok {
		if rel, ok := dest[dst]; ok {
			return rel, true
		}
	}

	if dest, ok := s.relationships[dst]; ok {
		if rel, ok := dest[src]; ok {
			return rel, true
		}
	}

	return nil, false
}

func (s *System) AddRelation(srcID, dstID, srcName, dstName string) (rv *Relation, ok bool) {
	src, ok := s.containers[SafeID(srcID)]
	if !ok {
		return nil, false
	}

	dst, ok := s.containers[SafeID(dstID)]
	if !ok {
		return nil, false
	}

	if rv, ok = s.findRelation(src.Name, dst.Name); ok {
		return rv, true
	}

	srcID, dstID = SafeID(src.Name), SafeID(dst.Name)

	dest, ok := s.relationships[srcID]
	if !ok {
		dest = make(map[string]*Relation)
	}

	rv, ok = dest[dstID]
	if !ok {
		rv = &Relation{
			Src: srcName,
			Dst: dstName,
		}
	}

	dest[dstID] = rv

	s.relationships[srcID] = dest

	return rv, true
}

func (s *System) WriteContainers(w io.Writer, level int) {
	putCommon(w, level, s.Description, "", s.Tags)

	next := level + 1

	for _, cID := range s.order {
		cont := s.containers[cID]

		putBlock(w, next, blockContainer, cont.ID, cont.Name)
		cont.Write(w, next+1)
		putEnd(w, next)
	}
}

func (s *System) WriteViews(w io.Writer, level int) {
	next := level + 1

	for _, cID := range s.order {
		cont := s.containers[cID]

		putView(w, level, blockComponent, cont.ID)

		putRaw(w, next, "include *")
		putRaw(w, next, "autoLayout")

		putEnd(w, level)
	}
}

func (s *System) WriteRelations(w io.Writer, level int) {
	for srcID, dest := range s.relationships {
		for dstID, rel := range dest {
			putRelation(w, level, srcID, dstID, rel.Tags)
			putEnd(w, level)
		}
	}
}
