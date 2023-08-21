package graph

import (
	"strconv"

	"github.com/s0rg/set"
)

type Container struct {
	Endpoints map[string]string
	ID        string
	Name      string
	Image     string
	outbounds []*Connection
	listeners []*Connection
}

func (ct *Container) SetConnections(conns []*Connection) {
	lseen := make(set.Set[string])
	oseen := make(set.Set[string])

	for _, con := range conns {
		switch {
		case con.IsListener():
			key := con.Kind.String() + strconv.Itoa(int(con.LocalPort))

			if lseen.TryAdd(key) {
				ct.listeners = append(ct.listeners, con)
			}
		case !con.IsInbound():
			key := con.RemoteIP.String() + con.Kind.String() + strconv.Itoa(int(con.RemotePort))

			if oseen.TryAdd(key) {
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
