package client_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/client"
)

func TestInodes(t *testing.T) {
	t.Parallel()

	m := &client.InodesMap{}

	if m.Has("1", 1, 1) {
		t.Fail()
	}

	// listener
	m.AddProcess("1", 1, "app1")
	m.AddInode("1", 1, 101)

	// client
	m.AddProcess("2", 2, "app2")
	m.MarkUnknown("2", 2, 101)
	m.MarkListener("1", 1, "/some/sock")

	m.ResolveUnknown(func(srcCID, dstCID, srcName, dstName, path string) {
		if srcCID != "1" || dstCID != "2" || srcName != "app1" || dstName != "app2" || path != "/some/sock" {
			t.Fail()
		}
	})

	if !m.Has("1", 1, 101) {
		t.Fail()
	}

	if m.Has("1", 2, 1) || m.Has("3", 1, 1) {
		t.Fail()
	}
}
