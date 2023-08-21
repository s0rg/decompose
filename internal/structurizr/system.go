package srtructurizr

import (
	"fmt"
	"io"
	"strings"
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

func (s *System) AddRelation(srcID, dstID string) (rv *Relation, ok bool) {
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
		rv = &Relation{}
	}

	dest[dstID] = rv

	s.relationships[srcID] = dest

	return rv, true
}

func (s *System) WriteContainers(w io.Writer) {
	const tabs = "\t\t\t"

	if s.Description != "" {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `description "%s"`, s.Description)
		fmt.Fprintln(w, "")
	}

	if len(s.Tags) > 0 {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `tags "%s"`, strings.Join(s.Tags, ","))
		fmt.Fprintln(w, "")
	}

	for _, cont := range s.containers {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, "%s = container ", cont.ID)
		fmt.Fprintf(w, `"%s" {`, cont.Name)
		fmt.Fprintln(w, "")
		cont.Write(w)
		fmt.Fprint(w, tabs)
		fmt.Fprintln(w, "}")
	}
}

func (s *System) WriteRelations(w io.Writer) {
	const tabs = "\t\t"

	for srcID, dest := range s.relationships {
		for dstID, rel := range dest {
			fmt.Fprint(w, tabs)
			fmt.Fprintf(w, "%s -> %s {\n", srcID, dstID)
			rel.Write(w)
			fmt.Fprint(w, tabs)
			fmt.Fprintln(w, "}")
		}
	}
}
