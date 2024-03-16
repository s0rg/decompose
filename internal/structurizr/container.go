package srtructurizr

import (
	"io"
)

type Container struct {
	ID          string
	Name        string
	Description string
	Technology  string
	Tags        []string
	Components  []*Component
}

func (c *Container) Write(w io.Writer, level int) {
	putCommon(w, level, c.Description, c.Technology, c.Tags)

	for _, com := range c.Components {
		putBlock(w, level, blockComponent, com.ID, com.Name)
		com.Write(w, level+1)
		putEnd(w, level)
	}
}
