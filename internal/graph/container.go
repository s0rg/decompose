package graph

import (
	"strconv"

	"github.com/s0rg/set"
)

type (
	ProcessInfo struct {
		Cmd []string
		Env []string
	}

	VolumeInfo struct {
		Type string
		Src  string
		Dst  string
	}

	Container struct {
		Endpoints map[string]string
		ID        string
		Name      string
		Image     string
		Process   *ProcessInfo
		Volumes   []*VolumeInfo
		outbounds []*Connection
		listeners []*Connection
	}
)

func (ct *Container) ConnectionsCount() int {
	return len(ct.outbounds) + len(ct.listeners)
}

func (ct *Container) SetConnections(conns []*Connection) {
	lseen := make(set.Unordered[string])
	oseen := make(set.Unordered[string])

	for _, con := range conns {
		switch {
		case con.IsListener():
			key := con.Kind.String() + strconv.Itoa(int(con.LocalPort))

			if lseen.Add(key) {
				ct.listeners = append(ct.listeners, con)
			}
		case !con.IsInbound():
			key := con.RemoteIP.String() + con.Kind.String() + strconv.Itoa(int(con.RemotePort))

			if oseen.Add(key) {
				ct.outbounds = append(ct.outbounds, con)
			}
		}
	}
}

func (ct *Container) ForEachOutbound(it func(*Connection)) {
	for _, con := range ct.outbounds {
		it(con)
	}
}

func (ct *Container) ForEachListener(it func(*Connection)) {
	for _, con := range ct.listeners {
		it(con)
	}
}
