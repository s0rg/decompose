package graph

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

type (
	MetaLoader struct {
		state map[string]*node.Meta
	}

	match struct {
		Key    string
		Weight int
	}
)

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

	matches := []*match{}

	for key := range ex.state {
		if !strings.HasPrefix(n.Name, key) {
			continue
		}

		matches = append(matches, &match{
			Key:    key,
			Weight: len(n.Name) - len(key),
		})
	}

	var meta *node.Meta

	switch len(matches) {
	default:
		slices.SortFunc(matches, func(a, b *match) int {
			return cmp.Compare(a.Weight, b.Weight)
		})

		fallthrough
	case 1:
		meta = ex.state[matches[0].Key]
	case 0:
		return
	}

	n.Meta = meta
}
