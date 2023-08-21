package srtructurizr

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
)

type Relation struct {
	Description string
	Technology  string
	Tags        []string
}

func (r *Relation) Write(w io.Writer) {
	const tabs = "\t\t\t"

	if r.Description != "" {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `description "%s"`, r.Description)
		fmt.Fprintln(w, "")
	}

	if r.Technology != "" {
		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `technology "%s"`, r.Technology)
		fmt.Fprintln(w, "")
	}

	if len(r.Tags) > 0 {
		sort.Strings(r.Tags)
		r.Tags = slices.Compact(r.Tags)

		fmt.Fprint(w, tabs)
		fmt.Fprintf(w, `tags "%s"`, strings.Join(r.Tags, ","))
		fmt.Fprintln(w, "")
	}
}
