package graph

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/s0rg/decompose/internal/node"
)

const (
	minItems  = 2
	minReport = 10
)

var ErrNotEnough = errors.New("not enough items")

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
	onlyLocal bool,
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

	if len(containers) < minItems {
		return fmt.Errorf("%w: containers", ErrNotEnough)
	}

	neighbours := buildIPMap(containers)

	log.Println("Building nodes...")

	nodes := buildNodes(containers, neighbours, follow, onlyLocal)

	log.Printf("Found %d nodes", len(nodes))

	if len(nodes) < minItems {
		return fmt.Errorf("%w: nodes", ErrNotEnough)
	}

	log.Println("Processing nodes...")

	for _, node := range nodes {
		node.Ports = node.Ports.Dedup()

		if err = grb.AddNode(node); err != nil {
			return fmt.Errorf("node '%s': %w", node.Name, err)
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
	onlyLocal bool,
) (rv map[string]*node.Node) {
	var skip bool

	rv = make(map[string]*node.Node)

	for _, con := range cntrs {
		skip = !con.Match(follow)

		n := &node.Node{ID: con.ID, Name: con.Name, Image: con.Image}

		con.ForEachOutbound(func(c *Connection) {
			rip := c.RemoteIP.String()
			rport := node.Port{Kind: c.Kind.String(), Value: int(c.RemotePort)}

			if lc, ok := neighbours[rip]; ok {
				if skip && lc.Match(follow) {
					skip = false
				}

				return
			}

			if onlyLocal || skip {
				return
			}

			rem, ok := rv[rip]
			if !ok {
				rem = &node.Node{ID: rip, Name: rip}
				rv[rip] = rem
			}

			rem.Ports = append(rem.Ports, rport)
		})

		if skip {
			continue
		}

		n.Networks = make([]string, 0, len(con.Endpoints))

		for _, epn := range con.Endpoints {
			n.Networks = append(n.Networks, epn)
		}

		con.ForEachListener(func(c *Connection) {
			n.Ports = append(n.Ports, node.Port{Kind: c.Kind.String(), Value: int(c.LocalPort)})
		})

		rv[con.ID] = n
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
			port := node.Port{Kind: c.Kind.String(), Value: int(c.RemotePort)}
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
