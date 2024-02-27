package graph

import (
	"cmp"
	"slices"

	"github.com/s0rg/set"
)

type ConnGroup struct {
	listenSeen set.Unordered[uint64]
	connSeen   set.Unordered[uint64]
	listen     []*Connection
	connected  []*Connection
}

func (cg *ConnGroup) Len() (rv int) {
	return len(cg.listen) + len(cg.connected)
}

func (cg *ConnGroup) AddListener(c *Connection) {
	cid, ok := c.UniqID()
	if !ok {
		return
	}

	if cg.listenSeen == nil {
		cg.listenSeen = make(set.Unordered[uint64])
	}

	if !cg.listenSeen.Add(cid) {
		return
	}

	cg.listen = append(cg.listen, c)
}

func (cg *ConnGroup) AddOutbound(c *Connection) {
	cid, ok := c.UniqID()
	if !ok {
		return
	}

	if cg.connSeen == nil {
		cg.connSeen = make(set.Unordered[uint64])
	}

	if !cg.connSeen.Add(cid) {
		return
	}

	cg.connected = append(cg.connected, c)
}

func (cg *ConnGroup) IterOutbounds(it func(*Connection)) {
	for _, con := range cg.connected {
		it(con)
	}
}

func (cg *ConnGroup) IterListeners(it func(*Connection)) {
	for _, con := range cg.listen {
		it(con)
	}
}

func (cg *ConnGroup) Sort() {
	slices.SortFunc(cg.listen, compare)
	slices.SortFunc(cg.connected, compare)
}

func compare(a, b *Connection) int {
	if a.Proto == b.Proto {
		return cmp.Compare(a.LocalPort, b.LocalPort)
	}

	return cmp.Compare(a.Proto, b.Proto)
}
