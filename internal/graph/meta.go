package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

type MetaLoader struct {
	state map[string]*node.Meta
}

func NewMetaLoader() *MetaLoader {
	return &MetaLoader{
		state: make(map[string]*node.Meta),
	}
}

func (ex *MetaLoader) FromReader(r io.Reader) error {
	d := json.NewDecoder(r)

	for d.More() {
		if err := d.Decode(&ex.state); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}

	return nil
}

func (ex *MetaLoader) Enrich(n *node.Node) {
	if len(ex.state) == 0 {
		return
	}

	for key, info := range ex.state {
		if !strings.HasPrefix(n.Name, key) {
			continue
		}

		n.Meta = info

		break
	}
}
