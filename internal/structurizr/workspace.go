package srtructurizr

import (
	"cmp"
	"fmt"
	"io"
	"slices"
)

type Workspace struct {
	relationships map[string]map[string]*Relation
	systems       map[string]*System
	Name          string
	Description   string
	defaultSystem string
	systemsOrder  []string
}

func NewWorkspace(name, defaultSystem string) *Workspace {
	return &Workspace{
		Name:          name,
		defaultSystem: defaultSystem,
		systemsOrder:  []string{SafeID(defaultSystem)},
		systems:       make(map[string]*System),
		relationships: make(map[string]map[string]*Relation),
	}
}

func (ws *Workspace) System(name string) (rv *System) {
	id := SafeID(name)

	if sys, ok := ws.systems[id]; ok {
		return sys
	}

	rv = NewSystem(name)

	ws.systems[rv.ID] = rv

	if rv.ID != ws.defaultSystem {
		ws.systemsOrder = append(ws.systemsOrder, rv.ID)
	}

	return rv
}

func (ws *Workspace) HasSystem(name string) (yes bool) {
	_, yes = ws.systems[SafeID(name)]

	return
}

func (ws *Workspace) AddRelation(srcID, dstID, srcName, dstName string) (rv *Relation, ok bool) {
	srcID, dstID = SafeID(srcID), SafeID(dstID)

	if !ws.HasSystem(srcID) || !ws.HasSystem(dstID) {
		return
	}

	dmap, ok := ws.relationships[srcID]
	if !ok {
		dmap = make(map[string]*Relation)
		ws.relationships[srcID] = dmap
	}

	if rv, ok = dmap[dstID]; ok {
		return rv, ok
	}

	rv = &Relation{
		Src: srcName,
		Dst: dstName,
	}

	dmap[dstID] = rv

	return rv, true
}

func (ws *Workspace) Write(w io.Writer) {
	var level int

	slices.Sort(ws.systemsOrder[1:])

	putHeader(w, level, headerWorkspace)

	level++
	putKey(w, level, keyName, ws.Name)
	putKey(w, level, keyDescription, ws.Description)

	fmt.Fprintln(w, "")
	putHeader(w, level, headerModel)

	level++

	for _, key := range ws.systemsOrder {
		system := ws.systems[key]

		putBlock(w, level, blockSystem, system.ID, system.Name)
		system.WriteContainers(w, level)
		putEnd(w, level)
	}

	fmt.Fprintln(w, "")

	for _, key := range ws.systemsOrder {
		system := ws.systems[key]

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
	relOrder := make([]string, 0, len(ws.relationships))

	for srcID := range ws.relationships {
		relOrder = append(relOrder, srcID)
	}

	slices.Sort(relOrder)

	for _, srcID := range relOrder {
		dest := ws.relationships[srcID]

		dstOrder := make([]string, 0, len(dest))

		for dstID := range dest {
			dstOrder = append(dstOrder, dstID)
		}

		slices.SortFunc(dstOrder, cmp.Compare)

		for _, dstID := range dstOrder {
			rel := dest[dstID]

			putRelation(w, level, srcID, dstID, rel.Tags)
			putEnd(w, level)
		}
	}
}

func (ws *Workspace) writeDefaultIncludes(w io.Writer, level int) {
	for _, id := range ws.systemsOrder {
		if id == ws.defaultSystem {
			continue
		}

		putRaw(w, level, "include "+id)
	}
}

func (ws *Workspace) writeViews(w io.Writer, level int) {
	putHeader(w, level, headerViews)

	level++

	for _, key := range ws.systemsOrder {
		system := ws.systems[key]

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

		system.WriteViews(w, level)
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
