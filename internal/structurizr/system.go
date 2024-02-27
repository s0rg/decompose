package srtructurizr

import (
	"cmp"
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
}

func NewSystem(name string) *System {
	return &System{
		ID:            safeID(name),
		Name:          name,
		containers:    make(map[string]*Container),
		relationships: make(map[string]map[string]*Relation),
	}
}

func (s *System) AddContainer(id, name string) (c *Container, ok bool) {
	id = safeID(id)

	if _, ok = s.containers[id]; ok {
		return nil, false
	}

	c = &Container{
		ID:   safeID(name),
		Name: name,
	}

	s.containers[id] = c

	return c, true
}

func (s *System) findRelation(src, dst string) (rv *Relation, found bool) {
	src, dst = safeID(src), safeID(dst)

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
	src, ok := s.containers[safeID(srcID)]
	if !ok {
		return nil, false
	}

	dst, ok := s.containers[safeID(dstID)]
	if !ok {
		return nil, false
	}

	if rv, ok = s.findRelation(src.Name, dst.Name); ok {
		return rv, true
	}

	srcID, dstID = safeID(src.Name), safeID(dst.Name)

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

	contOrder := make([]string, 0, len(s.containers))

	for cID := range s.containers {
		contOrder = append(contOrder, cID)
	}

	slices.SortFunc(contOrder, cmp.Compare)

	for _, cID := range contOrder {
		cont := s.containers[cID]

		putBlock(w, next, blockContainer, cont.ID, cont.Name)
		cont.Write(w, next+1)
		putEnd(w, next)
	}
}

func (s *System) WriteRelations(w io.Writer, level int) {
	for srcID, dest := range s.relationships {
		for dstID, rel := range dest {
			putRelation(w, level, srcID, dstID)
			rel.Write(w, level+1)
			putEnd(w, level)
		}
	}
}
