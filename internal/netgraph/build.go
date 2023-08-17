package netgraph

import (
	"context"
	"fmt"
	"log"

	"github.com/s0rg/decompose/internal/node"
)

const (
	minContainers = 2
	minReport     = 10
)

type ContainerClient interface {
	Containers(context.Context, NetProto, func(int, int)) ([]*Container, error)
}

type Builder interface {
	AddNode(*node.Node) error
	AddEdge(string, string, node.Port)
}

func Build(
	cli ContainerClient,
	grb Builder,
	knd NetProto,
	follow string,
	local bool,
) error {
	log.Println("Gathering containers info...")

	containers, err := cli.Containers(context.Background(), knd, func(cur, total int) {
		switch {
		case cur == 0:
			return
		case cur < total && cur%minReport > 0:
			return
		}

		log.Printf("Processing %d / %d [%.02f%%]", cur, total, percentOf(cur, total))
	})
	if err != nil {
		return fmt.Errorf("containers: %w", err)
	}

	log.Printf("Found %d alive containers", len(containers))

	if len(containers) < minContainers {
		log.Println("No suitable amount of containers found, nothing to do...")

		return nil
	}

	neighbours := buildIPMap(containers)

	log.Println("Building nodes...")

	nodes := buildNodes(containers, neighbours, follow, local)

	log.Printf("Found %d nodes", len(nodes))

	if len(nodes) < minContainers {
		log.Println("Not enought nodes found, nothing to do...")

		return nil
	}

	log.Println("Processing nodes...")

	for id, node := range nodes {
		node.Ports = node.Ports.Dedup()

		if err = grb.AddNode(node); err != nil {
			return fmt.Errorf("node [%s]: %w", id, err)
		}
	}

	log.Println("Building edges...")

	buildEdges(containers, neighbours, nodes, follow, grb.AddEdge)

	log.Println("Done!")

	return nil
}

func buildIPMap(cntrs []*Container) (rv map[string]*Container) {
	rv = make(map[string]*Container)

	for _, it := range cntrs {
		for ip := range it.Endpoints {
			rv[ip] = it
		}
	}

	return rv
}

func buildNodes(
	cntrs []*Container,
	neighbours map[string]*Container,
	follow string,
	local bool,
) (rv map[string]*node.Node) {
	rv = make(map[string]*node.Node)

	var skip bool

	for _, con := range cntrs {
		skip = !con.Match(follow)

		n := &node.Node{
			ID:    con.ID,
			Name:  con.Name,
			Image: con.Image,
		}

		con.ForEachListener(func(c *Connection) {
			n.Ports = append(n.Ports, node.Port{
				Value: int(c.LocalPort),
				Kind:  c.Kind.String(),
			})
		})

		con.ForEachOutbound(func(c *Connection) {
			rip := c.RemoteIP.String()

			rport := node.Port{
				Kind:  c.Kind.String(),
				Value: int(c.RemotePort),
			}

			if lc, ok := neighbours[rip]; ok {
				if skip && lc.Match(follow) {
					skip = false
				}

				return
			}

			if local {
				return
			}

			rem, ok := rv[rip]
			if !ok {
				rem = &node.Node{
					ID:   rip,
					Name: rip,
				}

				rv[rip] = rem
			}

			rem.Ports = append(rem.Ports, rport)
		})

		if !skip {
			rv[con.ID] = n
		}
	}

	return rv
}

func buildEdges(
	cntrs []*Container,
	local map[string]*Container,
	nodes map[string]*node.Node,
	follow string,
	edgefn func(src, dst string, port node.Port),
) {
	for _, con := range cntrs {
		src, ok := nodes[con.ID]
		if !ok {
			continue
		}

		con.ForEachOutbound(func(c *Connection) {
			port := node.Port{
				Kind:  c.Kind.String(),
				Value: int(c.RemotePort),
			}

			key := c.RemoteIP.String()

			if ldst, ok := local[key]; ok && (ldst.Match(follow) || con.Match(follow)) {
				edgefn(src.ID, ldst.ID, port)

				return
			}

			if !con.Match(follow) {
				return
			}

			if rdst, ok := nodes[key]; ok {
				edgefn(src.ID, rdst.ID, port)

				return
			}
		})
	}
}

func percentOf(a, b int) float64 {
	const hundred = 100.0

	return float64(a) / float64(b) * hundred
}
