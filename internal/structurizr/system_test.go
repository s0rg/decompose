package srtructurizr_test

import (
	"bytes"
	"strings"
	"testing"

	srtructurizr "github.com/s0rg/decompose/internal/structurizr"
)

func TestSystemRelation(t *testing.T) {
	t.Parallel()

	s := srtructurizr.NewSystem("")
	s.AddContainer("id1", "name1")
	s.AddContainer("id2", "name2")

	if _, ok := s.AddRelation("id1", "id2", "id1", "id2"); !ok {
		t.Fail()
	}

	if _, ok := s.AddRelation("id2", "id1", "id2", "id1"); !ok {
		t.Fail()
	}

	if _, ok := s.AddRelation("id2", "id1", "id2", "id1"); !ok {
		t.Fail()
	}

	if _, ok := s.AddRelation("id1", "id2", "id1", "id2"); !ok {
		t.Fail()
	}

	if _, ok := s.AddRelation("id1", "id3", "id1", "id3"); ok {
		t.Fail()
	}

	if _, ok := s.AddRelation("id3", "id1", "id3", "id1"); ok {
		t.Fail()
	}

	var b bytes.Buffer

	s.WriteRelations(&b, 0)

	if strings.Count(b.String(), "name1 -> name2") != 1 {
		t.Fail()
	}
}
