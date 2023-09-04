package srtructurizr

import (
	"fmt"
	"io"
)

type Workspace struct {
	systems       map[string]*System
	relationships map[string]map[string]*Relation
	Name          string
	Description   string
	defaultSystem string
}

func NewWorkspace(name, defaultSystem string) *Workspace {
	return &Workspace{
		Name:          name,
		defaultSystem: defaultSystem,
		systems:       make(map[string]*System),
		relationships: make(map[string]map[string]*Relation),
	}
}

func (ws *Workspace) System(name string) (rv *System) {
	id := safeID(name)

	if sys, ok := ws.systems[id]; ok {
		return sys
	}

	rv = NewSystem(name)

	ws.systems[rv.ID] = rv

	return rv
}

func (ws *Workspace) HasSystem(name string) (yes bool) {
	_, yes = ws.systems[safeID(name)]

	return
}

func (ws *Workspace) AddRelation(srcID, dstID string) (rv *Relation, ok bool) {
	srcID, dstID = safeID(srcID), safeID(dstID)

	dmap, ok := ws.relationships[srcID]
	if !ok {
		dmap = make(map[string]*Relation)
		ws.relationships[srcID] = dmap
	}

	if rv, ok = dmap[dstID]; ok {
		return rv, ok
	}

	rv = &Relation{}

	dmap[dstID] = rv

	return rv, true
}

func (ws *Workspace) Write(w io.Writer) {
	var level int

	putHeader(w, level, headerWorkspace)

	level++
	putKey(w, level, keyName, ws.Name)
	putKey(w, level, keyDescription, ws.Description)

	fmt.Fprintln(w, "")
	putHeader(w, level, headerModel)

	level++

	for _, system := range ws.systems {
		putBlock(w, level, blockSystem, system.ID, system.Name)
		system.WriteContainers(w, level)
		putEnd(w, level)
	}

	fmt.Fprintln(w, "")

	for _, system := range ws.systems {
		system.WriteRelations(w, level)
	}

	ws.writeRelations(w, level)

	level--
	putEnd(w, level) // model

	fmt.Fprintln(w, "")

	ws.writeViews(w, level)

	level--
	putEnd(w, level) // workspace
}

func (ws *Workspace) writeRelations(w io.Writer, level int) {
	for srcID, dest := range ws.relationships {
		for dstID, rel := range dest {
			putRelation(w, level, srcID, dstID)
			rel.Write(w, level+1)
			putEnd(w, level)
		}
	}
}

func (ws *Workspace) writeDefaultIncludes(w io.Writer, level int) {
	for id := range ws.systems {
		if id == ws.defaultSystem {
			continue
		}

		putRaw(w, level, "include "+id)
	}
}

func (ws *Workspace) writeViews(w io.Writer, level int) {
	putHeader(w, level, headerViews)
	level++

	for _, system := range ws.systems {
		putView(w, level, blockSystemCtx, system.ID)
		level++

		putRaw(w, level, "include *")

		if system.Name == ws.defaultSystem {
			ws.writeDefaultIncludes(w, level)
		}

		putRaw(w, level, "autoLayout")

		level--
		putEnd(w, level) // system context

		putView(w, level, blockContainer, system.ID)
		level++

		putRaw(w, level, "include *")
		putRaw(w, level, "autoLayout")

		level--
		putEnd(w, level) // container
	}

	fmt.Fprintln(w, "")
	putHeader(w, level, "styles")
	level++

	putRaw(w, level, `element "Element" {`)
	level++

	putRaw(w, level, "metadata true")
	putRaw(w, level, "description true")

	level--
	putEnd(w, level) // element

	level--
	putEnd(w, level) // styles

	level--
	putEnd(w, level) // views
}
