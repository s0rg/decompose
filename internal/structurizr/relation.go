package srtructurizr

import (
	"io"
)

type Relation struct {
	Description string
	Technology  string
	Tags        []string
}

func (r *Relation) Write(w io.Writer, level int) {
	putCommon(w, level, r.Description, r.Technology, r.Tags)
}
