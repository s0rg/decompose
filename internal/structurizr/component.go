package srtructurizr

import "io"

type Component struct {
	ID          string
	Name        string
	Description string
	Technology  string
	Tags        []string
}

func (c *Component) Write(w io.Writer, level int) {
	putCommon(w, level, c.Description, c.Technology, c.Tags)
}
