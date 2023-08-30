package graph

import (
	"net"
)

type Connection struct {
	LocalIP    net.IP
	RemoteIP   net.IP
	LocalPort  uint16
	RemotePort uint16
	Proto      NetProto
}

func (cn *Connection) IsListener() bool {
	return cn.RemotePort == 0
}

func (cn *Connection) IsInbound() bool {
	return cn.LocalPort < cn.RemotePort
}
