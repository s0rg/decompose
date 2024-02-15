package graph

import (
	"cmp"
	"slices"

	"github.com/s0rg/set"
)

type connGroup struct {
	listenSeen set.Unordered[uint64]
	connSeen   set.Unordered[uint64]
	listen     []*Connection
	connected  []*Connection
}

func (cg *connGroup) Len() (rv int) {
	return len(cg.listen) + len(cg.connected)
}

func (cg *connGroup) AddListener(c *Connection) {
	cid, ok := c.UniqID()
	if !ok {
		return
	}

	if cg.listenSeen == nil {
		cg.listenSeen = make(set.Unordered[uint64])
	}

	if cg.listenSeen.Has(cid) {
		return
	}

	cg.listen = append(cg.listen, c)
}

func (cg *connGroup) AddOutbound(c *Connection) {
	cid, ok := c.UniqID()
	if !ok {
		return
	}

	if cg.connSeen == nil {
		cg.connSeen = make(set.Unordered[uint64])
	}

	if cg.connSeen.Has(cid) {
		return
	}

	cg.connected = append(cg.connected, c)
}

func (cg *connGroup) IterOutbounds(it func(*Connection)) {
	for _, con := range cg.connected {
		it(con)
	}
}

func (cg *connGroup) IterListeners(it func(*Connection)) {
	for _, con := range cg.listen {
		it(con)
	}
}

func (cg *connGroup) Sort() {
	slices.SortFunc(cg.listen, compare)
	slices.SortFunc(cg.connected, compare)
}

func compare(a, b *Connection) int {
	if a.Proto == b.Proto {
		return cmp.Compare(a.LocalPort, b.LocalPort)
	}

	return cmp.Compare(a.Proto, b.Proto)
}
