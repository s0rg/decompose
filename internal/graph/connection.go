package graph

import (
	"hash/fnv"
	"io"
	"net"
	"strconv"
)

type Connection struct {
	Process    string
	LocalIP    net.IP
	RemoteIP   net.IP
	RemotePort uint16
	LocalPort  uint16
	Proto      NetProto
}

func (c *Connection) IsListener() bool {
	return c.RemotePort == 0
}

func (c *Connection) IsInbound() bool {
	return c.LocalPort < c.RemotePort
}

func (c *Connection) UniqID() (id uint64, ok bool) {
	var key string

	switch {
	case c.IsListener():
		key, ok = c.Proto.String()+strconv.Itoa(int(c.LocalPort)), true
	case !c.IsInbound():
		key, ok = c.RemoteIP.String()+c.Proto.String()+strconv.Itoa(int(c.RemotePort)), true
	}

	if !ok {
		return
	}

	h := fnv.New64a()
	if _, err := io.WriteString(h, c.Process+key); err != nil {
		return 0, false
	}

	return h.Sum64(), true
}
