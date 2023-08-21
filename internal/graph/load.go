package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

const idSuffix = "-id"

type Loader struct {
	nodes map[string]*node.Node
	edges map[string]map[string]node.Ports
	cfg   *Config
}

func NewLoader(cfg *Config) *Loader {
	return &Loader{
		cfg:   cfg,
		nodes: make(map[string]*node.Node),
		edges: make(map[string]map[string]node.Ports),
	}
}

func (l *Loader) LoadStream(r io.Reader) error {
	jr := json.NewDecoder(r)

	for jr.More() {
		var n node.JSON

		if err := jr.Decode(&n); err != nil {
			return fmt.Errorf("decode: %w", err)
		}

		l.insert(&n)
	}

	return nil
}

func (l *Loader) prepareNode(n *node.JSON) (id string, rv *node.Node) {
	id = n.Name
	if !n.IsExternal {
		id += idSuffix
	}

	nod, ok := l.nodes[id]
	if !ok {
		nod = &node.Node{
			ID:       id,
			Name:     n.Name,
			Networks: n.Networks,
		}

		if n.Image != nil {
			nod.Image = *n.Image
		}

		if n.Meta != nil {
			nod.Meta = n.Meta
		} else {
			l.cfg.Enricher.Enrich(nod)
		}
	}

	if !(n.IsExternal && l.cfg.OnlyLocal) {
		nod.Ports = append(nod.Ports, l.preparePorts(n.Listen)...)
	}

	return id, nod
}

func (l *Loader) prepareEdges(id string, n *node.JSON) (rv map[string]node.Ports, skip bool) {
	var ok bool

	if rv, ok = l.edges[id]; !ok {
		rv = make(map[string]node.Ports)
	}

	skip = !l.cfg.MatchName(n.Name)

	for k, p := range n.Connected {
		prep := l.preparePorts(p)
		if len(prep) == 0 {
			continue
		}

		if skip && l.cfg.MatchName(k) {
			skip = false
		}

		if skip {
			continue
		}

		ports, ok := rv[k]
		if !ok {
			ports = []node.Port{}
		}

		rv[k] = append(ports, prep...)
	}

	return rv, skip
}

func (l *Loader) insert(n *node.JSON) {
	id, nod := l.prepareNode(n)
	cons, skip := l.prepareEdges(id, n)

	if skip {
		return
	}

	l.nodes[id] = nod
	l.edges[id] = cons
}

func (l *Loader) isExternalNode(id string) (yes bool) {
	n, ok := l.nodes[id]
	if !ok {
		return false
	}

	return n.IsExternal()
}

func (l *Loader) Build() error {
	for id, node := range l.nodes {
		if l.cfg.OnlyLocal && l.isExternalNode(id) {
			continue
		}

		if err := l.cfg.Builder.AddNode(node); err != nil {
			return fmt.Errorf("node %s: %w", node.Name, err)
		}
	}

	for srcID, dmap := range l.edges {
		if l.cfg.OnlyLocal && l.isExternalNode(srcID) {
			continue
		}

		for dstID, ports := range dmap {
			if l.isExternalNode(dstID) {
				if l.cfg.OnlyLocal {
					continue
				}
			} else {
				dstID += idSuffix
			}

			ports = ports.Dedup()

			for i := 0; i < len(ports); i++ {
				if l.cfg.MatchProto(ports[i].Kind) {
					l.cfg.Builder.AddEdge(srcID, dstID, ports[i])
				}
			}
		}
	}

	return nil
}

func (l *Loader) preparePorts(lst []string) (rv []node.Port) {
	rv = make([]node.Port, 0, len(lst))

	for _, v := range lst {
		if p, ok := parsePort(v); ok && l.cfg.MatchProto(p.Kind) {
			rv = append(rv, p)
		}
	}

	return slices.Clip(rv)
}

func parsePort(v string) (p node.Port, ok bool) {
	const (
		nparts = 2
		sep    = "/"
	)

	parts := strings.SplitN(v, sep, nparts)
	if len(parts) != nparts {
		return
	}

	iport, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}

	p.Value = iport
	p.Kind = parts[1]

	return p, true
}
