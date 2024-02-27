package srtructurizr

import (
	"fmt"
	"io"
)

type Relation struct {
	Src         string
	Dst         string
	Description string
	Technology  string
	Tags        []string
}

func (r *Relation) Write(w io.Writer, level int) {
	desc := fmt.Sprintf("%s to %s %s", r.Src, r.Dst, r.Description)

	putCommon(w, level, desc, r.Technology, r.Tags)
}
