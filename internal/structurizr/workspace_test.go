package srtructurizr_test

import (
	"bytes"
	"strings"
	"testing"

	srtructurizr "github.com/s0rg/decompose/internal/structurizr"
)

func TestWorkspaceRelation(t *testing.T) {
	t.Parallel()

	ws := srtructurizr.NewWorkspace("test", "1")

	if _, ok := ws.AddRelation("foo", "bar"); ok {
		t.Fail()
	}

	ws.System("1")
	s := ws.System("2")

	s.Tags = append(s.Tags, "")

	if _, ok := ws.AddRelation("1", "2"); !ok {
		t.Fail()
	}

	if _, ok := ws.AddRelation("2", "1"); !ok {
		t.Fail()
	}

	if _, ok := ws.AddRelation("1", "2"); !ok {
		t.Fail()
	}

	var b bytes.Buffer

	ws.Write(&b)

	if strings.Contains(b.String(), "tags") {
		t.Fail()
	}
}
