package builder_test

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/s0rg/decompose/internal/node"
)

const (
	goldenRoot = "testdata"
	goldenExt  = ".golden"
)

var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.MkdirAll(goldenRoot, 0o664)
	os.Exit(m.Run())
}

func golden(t *testing.T, file, actual string) string {
	t.Helper()

	path := filepath.Join(goldenRoot, file+goldenExt)

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("Error opening file %s: %s", path, err)
	}

	defer fd.Close()

	if *update {
		if _, werr := fd.WriteString(actual); err != nil {
			t.Fatalf("Error writing to file %s: %s", path, werr)
		}

		return actual
	}

	content, err := io.ReadAll(fd)
	if err != nil {
		t.Fatalf("Error reading file %s: %s", path, err)
	}

	return string(content)
}

func makeTestPorts(vals ...*node.Port) (rv *node.Ports) {
	rv = &node.Ports{}

	for _, p := range vals {
		rv.Add("", p)
	}

	return rv
}
