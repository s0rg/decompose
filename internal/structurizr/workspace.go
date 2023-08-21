package srtructurizr

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Workspace struct {
	System      *System
	Name        string
	Description string
}

func (ws *Workspace) Write(w io.Writer) {
	fmt.Fprintln(w, "workspace {")

	if ws.Name != "" {
		fmt.Fprintln(w, "\tname ")
		fmt.Fprintf(w, `"%s"`, ws.Name)
		fmt.Fprintln(w, "")
	}

	if ws.Description != "" {
		fmt.Fprintln(w, "\tdescription ")
		fmt.Fprintf(w, `"%s"`, ws.Description)
		fmt.Fprintln(w, "")
	}

	fmt.Fprintln(w, "\tmodel {")

	fmt.Fprintf(w, "\t\t%s = softwareSystem ", ws.System.ID)
	fmt.Fprintf(w, `"%s"`, ws.System.Name)
	fmt.Fprintln(w, " {")
	ws.System.WriteContainers(w)
	fmt.Fprintln(w, "\t\t}")
	ws.System.WriteRelations(w)
	fmt.Fprintln(w, "\t}\n\n\tviews {")
	fmt.Fprintf(w, "\t\tcontainer %s {\n", ws.System.ID)
	fmt.Fprintln(w, "\t\t\tinclude *\n\t\t\tautoLayout lr\n\t\t}\n\t}\n}")
}

func safeID(v string) (id string) {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '-' {
			return '_'
		}

		return r
	}, v)
}
