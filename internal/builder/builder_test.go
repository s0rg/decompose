package builder_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/builder"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	if _, ok := builder.Create(""); ok {
		t.Fail()
	}

	for _, name := range builder.Names() {
		if _, ok := builder.Create(name); !ok {
			t.Fail()
		}
	}
}

func TestSupportCluster(t *testing.T) {
	t.Parallel()

	does := []string{
		builder.KindDOT,
		builder.KindSTAT,
		builder.KindStructurizr,
		builder.KindPlantUML,
	}

	doesnt := []string{
		builder.KindJSON,
		builder.KindTREE,
		builder.KindYAML,
	}

	for _, k := range does {
		if !builder.SupportCluster(k) {
			t.Fail()
		}
	}

	for _, k := range doesnt {
		if builder.SupportCluster(k) {
			t.Fail()
		}
	}
}
