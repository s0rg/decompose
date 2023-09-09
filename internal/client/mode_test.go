package client_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/client"
)

func TestModeNone(t *testing.T) {
	t.Parallel()

	if client.None.String() != "" {
		t.Fail()
	}
}
