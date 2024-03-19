package graph

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/s0rg/decompose/internal/node"
)

const idSuffix = "-id"

type Loader struct {
	nodes map[string]*node.Node
	edges map[string]map[string][]*node.Connection
	cfg   *Config
}

func NewLoader(cfg *Config) *Loader {
	return &Loader{
		cfg:   cfg,
		nodes: make(map[string]*node.Node),
		edges: make(map[string]map[string][]*node.Connection),
	}
}

func (l *Loader) FromReader(r io.Reader) error {
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

func (l *Loader) Build() error {
	for id, node := range l.nodes {
		if l.cfg.OnlyLocal && l.isExternal(id) {
			continue
		}

		if err := l.cfg.Builder.AddNode(node); err != nil {
			return fmt.Errorf("node %s: %w", node.Name, err)
		}
	}

	for srcID, dmap := range l.edges {
		if l.cfg.OnlyLocal && l.isExternal(srcID) {
			continue
		}

		l.connect(srcID, dmap)
	}

	return nil
}

func (l *Loader) createNode(id string, n *node.JSON) (rv *node.Node) {
	rv = &node.Node{
		ID:        id,
		Name:      n.Name,
		Container: n.Container,
		Ports:     &node.Ports{},
		Networks:  []string{},
	}

	if n.Image != nil {
		rv.Image = *n.Image
	}

	if len(n.Tags) > 0 {
		rv.Meta = &node.Meta{
			Tags: n.Tags,
		}
	}

	if len(n.Networks) > 0 {
		rv.Networks = n.Networks
	}

	if !l.cfg.FullInfo {
		return rv
	}

	if len(n.Volumes) > 0 {
		rv.Volumes = n.Volumes
	}

	return rv
}

func (l *Loader) loadNode(n *node.JSON) (id string, rv *node.Node) {
	id = n.Name
	if !n.IsExternal {
		id += idSuffix
	}

	nod, ok := l.nodes[id]
	if !ok {
		nod = l.createNode(id, n)
	}

	if !(n.IsExternal && l.cfg.OnlyLocal) {
		loadListeners(nod.Ports, n.Listen)
	}

	return id, nod
}

func (l *Loader) loadEdges(id string, n *node.JSON) (rv map[string][]*node.Connection, skip bool) {
	var ok bool

	if rv, ok = l.edges[id]; !ok {
		rv = make(map[string][]*node.Connection)
	}

	skip = !l.cfg.MatchName(n.Name)

	for k, p := range n.Connected {
		if l.cfg.NoLoops && n.Name == k {
			continue
		}

		if !skip || l.cfg.MatchName(k) {
			rv[k] = append(rv[k], p...)
			skip = false
		}
	}

	return rv, skip
}

func (l *Loader) insert(n *node.JSON) {
	id, nod := l.loadNode(n)
	cons, skip := l.loadEdges(id, n)

	if skip {
		return
	}

	l.cfg.Meta.Enrich(nod)

	nod.Ports.Compact()

	l.nodes[id] = nod
	l.edges[id] = cons
}

func (l *Loader) isExternal(id string) (yes bool) {
	n, ok := l.nodes[id]
	if !ok {
		return false
	}

	return n.IsExternal()
}

func (l *Loader) connect(
	srcID string,
	conns map[string][]*node.Connection,
) {
	for dstID, cl := range conns {
		if l.isExternal(dstID) {
			if l.cfg.OnlyLocal {
				continue
			}
		} else {
			dstID += idSuffix
		}

		for _, c := range cl {
			if !l.cfg.MatchProto(c.Port.Kind) {
				continue
			}

			l.cfg.Builder.AddEdge(&node.Edge{
				SrcID:   srcID,
				DstID:   dstID,
				SrcName: c.Src,
				DstName: c.Dst,
				Port:    c.Port,
			})
		}
	}
}

func loadListeners(
	ports *node.Ports,
	conns map[string][]*node.Port,
) {
	for k, cl := range conns {
		for _, p := range cl {
			ports.Add(k, p)
		}
	}
}
