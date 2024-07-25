package graph

import (
	"hash/fnv"
	"io"
	"net"
	"strconv"
)

type Connection struct {
	Process string
	DstID   string
	Path    string
	SrcIP   net.IP
	DstIP   net.IP
	Inode   uint64
	SrcPort int
	DstPort int
	Proto   NetProto
	Listen  bool
}

func (c *Connection) IsListener() bool {
	return c.Listen
}

func (c *Connection) IsInbound() bool {
	if c.Proto == UNIX {
		return c.Listen
	}

	return c.SrcPort < c.DstPort
}

func (c *Connection) IsLocal() bool {
	if c.Proto == UNIX {
		return false
	}

	return c.SrcIP.IsLoopback()
}

func (c *Connection) UniqID() (id uint64, ok bool) {
	var key string

	switch {
	case c.Proto == UNIX:
		key = c.Path
	case c.IsListener():
		key = c.Proto.String() + strconv.Itoa(c.SrcPort)
	case !c.IsInbound():
		key = c.DstIP.String() + c.Proto.String() + strconv.Itoa(c.DstPort)
	default:
		return
	}

	h := fnv.New64a()
	_, _ = io.WriteString(h, c.Process+key)

	return h.Sum64(), true
}
