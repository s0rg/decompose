package graph

import (
	"fmt"
	"log"
	"strconv"

	"github.com/s0rg/decompose/internal/node"
)

type builderState struct {
	Config     *Config
	KnownIP    map[string]*Container
	Nodes      map[string]*node.Node
	Containers []*Container
}

func newBuilderState(
	cfg *Config,
	cntrs []*Container,
) (bs *builderState) {
	bs = &builderState{
		Config:     cfg,
		Containers: cntrs,
		KnownIP:    make(map[string]*Container, len(cntrs)),
		Nodes:      make(map[string]*node.Node, len(cntrs)),
	}

	for _, it := range cntrs {
		for ip := range it.Endpoints {
			bs.KnownIP[ip] = it
		}
	}

	return bs
}

func (bs *builderState) BuildNodes() (total int, err error) {
	var notice bool

	for _, con := range bs.Containers {
		if con.ConnectionsCount() == 0 && !notice {
			log.Printf("No connections for container: %s:%s, try run as root", con.ID, con.Name)

			notice = true
		}

		if !bs.matchContainer(con) {
			continue
		}

		n := con.ToNode()

		bs.Config.Meta.Enrich(n)

		if err = bs.Config.Builder.AddNode(n); err != nil {
			return 0, fmt.Errorf("node '%s': %w", n.Name, err)
		}

		bs.Nodes[con.ID] = n

		total++
	}

	return total, nil
}

func (bs *builderState) BuildEdges() (total int) {
	for _, con := range bs.Containers {
		src, ok := bs.Nodes[con.ID]
		if !ok {
			continue
		}

		con.IterOutbounds(func(c *Connection) {
			if edge, ok := bs.findEdge(con.ID, con.Name, c); ok {
				edge.SrcID = src.ID

				bs.Config.Builder.AddEdge(edge)

				total++
			}
		})
	}

	return total
}

func (bs *builderState) matchContainer(cn *Container) (yes bool) {
	yes = bs.Config.MatchName(cn.Name)

	cn.IterOutbounds(func(c *Connection) {
		if c.Proto == UNIX || c.DstIP.IsLoopback() {
			return
		}

		rip := c.DstIP.String()

		if lc, ok := bs.KnownIP[rip]; ok { // destination known
			if !yes && bs.Config.MatchName(lc.Name) {
				yes = true
			}

			return
		}

		if bs.Config.OnlyLocal || !yes {
			return
		}

		// destination is remote host, add it
		rem, ok := bs.Nodes[rip]
		if !ok {
			rem = node.External(rip)
			bs.Nodes[rip] = rem
		}

		rem.Ports.Add(ProcessRemote, &node.Port{
			Kind:   c.Proto.String(),
			Value:  strconv.Itoa(c.DstPort),
			Number: c.DstPort,
		})
	})

	return yes
}

func (bs *builderState) findEdge(cid, cname string, conn *Connection) (rv *node.Edge, ok bool) {
	var (
		port = &node.Port{
			Kind: conn.Proto.String(),
		}
		key string
	)

	rv = &node.Edge{
		SrcName: conn.Process,
		Port:    port,
	}

	switch conn.Proto {
	case UNIX:
		port.Value = conn.Path
		rv.DstID = conn.DstID
	default:
		key = conn.DstIP.String()
		port.Value = strconv.Itoa(conn.DstPort)
		port.Number = conn.DstPort

		if conn.DstIP.IsLoopback() {
			rv.DstID = cid
		} else if ldst, found := bs.KnownIP[key]; found {
			rv.DstID = ldst.ID
		}
	}

	if rv.DstID != "" {
		if bs.Config.NoLoops && cid == rv.DstID {
			return nil, false
		}

		dst, found := bs.Nodes[rv.DstID]
		if !found {
			return nil, false
		}

		dname, found := dst.Ports.Get(port)
		if !found {
			dname = ProcessUnknown
		}

		rv.DstName = dname

		return rv, true
	}

	if !bs.Config.MatchName(cname) {
		return nil, false
	}

	if rdst, found := bs.Nodes[key]; found {
		rv.DstID = rdst.ID
		rv.DstName = ProcessRemote

		return rv, true
	}

	return nil, false
}
