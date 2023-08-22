package srtructurizr

import (
	"fmt"
	"io"
)

type Workspace struct {
	System      *System
	Name        string
	Description string
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
	putBlock(w, level, blockSystem, ws.System.ID, ws.System.Name)
	ws.System.WriteContainers(w, level)
	putEnd(w, level)

	fmt.Fprintln(w, "")

	ws.System.WriteRelations(w, level)

	level--
	putEnd(w, level) // model

	fmt.Fprintln(w, "")

	putHeader(w, level, headerViews)

	level++
	putView(w, level, blockContainer, ws.System.ID)

	level++
	putRaw(w, level, "include *")
	putRaw(w, level, "autoLayout lr")

	level--
	putEnd(w, level) // container

	level--
	putEnd(w, level) // views

	level--
	putEnd(w, level) // workspace
}
