package builder

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"slices"

	"github.com/s0rg/decompose/internal/node"
)

type PlantUML struct {
	nodes map[string]*node.Node
	conns map[string]map[string][]*node.Port
	order []string
}

func NewPlantUML() *PlantUML {
	return &PlantUML{
		nodes: make(map[string]*node.Node),
		conns: make(map[string]map[string][]*node.Port),
	}
}

func (p *PlantUML) Name() string {
	return "plant-uml"
}

func (p *PlantUML) AddNode(n *node.Node) error {
	p.nodes[n.ID] = n
	p.order = append(p.order, n.ID)

	return nil
}

func (p *PlantUML) AddEdge(e *node.Edge) {
	nsrc, ok := p.nodes[e.SrcID]
	if !ok {
		return
	}

	ndst, ok := p.nodes[e.DstID]
	if !ok {
		return
	}

	if !e.Port.Local && nsrc.Cluster != ndst.Cluster {
		e.SrcID, e.DstID = nsrc.Cluster, ndst.Cluster
	}

	mdst, ok := p.conns[e.SrcID]
	if !ok {
		mdst = make(map[string][]*node.Port)
		p.conns[e.SrcID] = mdst
	}

	ports, ok := mdst[e.DstID]
	if !ok {
		ports = make([]*node.Port, 0, 1)
	}

	mdst[e.DstID] = append(ports, e.Port)
}

func (p *PlantUML) Write(w io.Writer) error {
	fmt.Fprintln(w, "@startuml")
	fmt.Fprintln(w, "skinparam componentStyle rectangle")
	fmt.Fprintln(w, "skinparam nodesep 5")
	fmt.Fprintln(w, "skinparam ranksep 5")

	p.writeNodes(w)

	fmt.Fprintln(w, "")

	p.writeEdges(w)

	fmt.Fprintln(w, "@enduml")

	return nil
}

func (p *PlantUML) writeNodes(w io.Writer) {
	slices.Sort(p.order)

	cloud := []*node.Node{}

	for _, id := range p.order {
		nod := p.nodes[id]

		if nod.IsExternal() {
			cloud = append(cloud, nod)

			continue
		}

		np := []*node.Port{}

		fmt.Fprintf(w, "component \"%s\" as %s {\n", nod.Name,
			makeID(nod.Cluster, nod.Name),
		)

		nod.Ports.Iter(func(process string, ports []*node.Port) {
			fmt.Fprintf(w, " component \"%s\" as %s {\n",
				process,
				makeID(nod.Cluster, nod.Name, process),
			)

			for _, prt := range ports {
				fmt.Fprintf(w, "  portin \"%s\" as %s\n",
					prt.Label(),
					makeID(nod.Cluster, nod.Name, process, prt.Label()),
				)

				if !prt.Local {
					np = append(np, prt)
				}
			}

			fmt.Fprintln(w, " }")
		})

		for _, prt := range np {
			fmt.Fprintf(w, " portin \"%s\" as %s\n",
				prt.Label(),
				makeID(nod.Cluster, nod.Name, prt.Label()),
			)
		}

		fmt.Fprintln(w, "}")
	}

	if len(cloud) > 0 {
		fmt.Fprintln(w, "cloud \"Externals\" as ext {")

		for _, nod := range cloud {
			fmt.Fprintf(w, " component \"%s\" as %s {\n", nod.Name, makeID("ext", nod.Name))

			nod.Ports.Iter(func(_ string, ports []*node.Port) {
				for _, prt := range ports {
					fmt.Fprintf(w, "  portin \"%s\" as %s\n",
						prt.Label(),
						makeID("ext", nod.Name, prt.Label()),
					)
				}

				fmt.Fprintln(w, "  }")
			})

			fmt.Fprintln(w, " }")
		}
	}

	fmt.Fprintln(w, "}")
}

func (p *PlantUML) writeEdges(w io.Writer) {
	order := make([]string, 0, len(p.conns))

	for k := range p.conns {
		order = append(order, k)
	}

	slices.Sort(order)

	for _, src := range order {
		dmap := p.conns[src]

		nsrc, ok := p.nodes[src]
		if !ok {
			continue
		}

		locals := make(map[string]string)

		nsrc.Ports.Iter(func(process string, ports []*node.Port) {
			for _, prt := range ports {
				locals[prt.Label()] = process
			}
		})

		nsrc.Ports.Iter(func(process string, ports []*node.Port) {
			for _, prt := range ports {
				if !prt.Local {
					fmt.Fprintf(w, "%s -> %s\n",
						makeID(nsrc.Cluster, nsrc.Name, prt.Label()),
						makeID(nsrc.Cluster, nsrc.Name, process, prt.Label()),
					)
				}
			}
		})

		dorder := make([]string, 0, len(dmap))

		for k := range dmap {
			dorder = append(dorder, k)
		}

		slices.Sort(dorder)

		for _, dst := range dorder {
			ports := dmap[dst]
			ndst := p.nodes[dst]

			for _, prt := range ports {
				if prt.Local {
					dstp := locals[prt.Label()]

					fmt.Fprintf(w, "%s --> %s\n",
						makeID(nsrc.Cluster, nsrc.Name),
						makeID(nsrc.Cluster, nsrc.Name, dstp, prt.Label()),
					)
				} else {
					fmt.Fprintf(w, "%s -----> %s: %s\n",
						makeID(nsrc.Cluster, nsrc.Name),
						makeID(ndst.Cluster, ndst.Name, prt.Label()),
						prt.Label(),
					)
				}
			}
		}
	}
}

func makeID(parts ...string) (rv string) {
	h := fnv.New64a()

	for _, p := range parts {
		_, _ = io.WriteString(h, p)
	}

	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, h.Sum64())

	return "id_" + hex.EncodeToString(b[:n])
}
