package builder

import (
	"slices"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

func joinConnections(conns []*node.Connection, sep string) (rv string) {
	raw := make([]string, 0, len(conns))

	for _, c := range conns {
		raw = append(raw, c.Port.Label())
	}

	slices.Sort(raw)

	return strings.Join(raw, sep)
}

func joinListeners(ports map[string][]*node.Port, sep string) (rv string) {
	var tmp []string

	for _, plist := range ports {
		for _, p := range plist {
			tmp = append(tmp, p.Label())
		}
	}

	slices.Sort(tmp)

	return strings.Join(tmp, sep)
}
