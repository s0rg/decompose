package srtructurizr

import (
	"fmt"
	"io"
	"strings"
)

type Container struct {
	ID          string
	Name        string
	Description string
	Technology  string
	Tags        []string
}

func (c *Container) Write(w io.Writer) {
	const tabs = "\t\t\t\t"

	if c.Description != "" {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `description "%s"`, c.Description)
		fmt.Fprintln(w, "")
	}

	if c.Technology != "" {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `technology "%s"`, c.Technology)
		fmt.Fprintln(w, "")
	}

	if len(c.Tags) > 0 {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `tags "%s"`, strings.Join(c.Tags, ","))
		fmt.Fprintln(w, "")
	}
}
