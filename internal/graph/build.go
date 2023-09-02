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
	Containers(context.Context, NetProto, bool, func(int, int)) ([]*Container, error)
}

type Builder interface {
	AddNode(*node.Node) error
	AddEdge(string, string, node.Port)
}

type Enricher interface {
	Enrich(*node.Node)
}

type Assigner interface {
	Assign(*node.Node)
	IsEmpty() bool
}

func Build(
	cfg *Config,
	cli ContainerClient,
) error {
	log.Println("Gathering containers info")

	containers, err := cli.Containers(
		context.Background(),
		cfg.Proto,
		cfg.FullInfo,
		func(cur, total int) {
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

	log.Println("Building nodes")

	nodes := createNodes(cfg, containers, neighbours)

	log.Printf("Processing %d nodes", len(nodes))

	if len(nodes) < minItems {
		return fmt.Errorf("%w: nodes", ErrNotEnough)
	}

	for _, node := range nodes {
		node.Ports = node.Ports.Dedup()

		cfg.Meta.Enrich(node)
		cfg.Cluster.Assign(node)

		if err = cfg.Builder.AddNode(node); err != nil {
			return fmt.Errorf("node '%s': %w", node.Name, err)
		}
	}

	log.Println("Building edges")

	if cfg.Cluster.IsEmpty() {
		directEdges(cfg, containers, neighbours, nodes)
	} else {
		clusterEdges(cfg, containers, neighbours, nodes)
	}

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

func createNodes(
	cfg *Config,
	cntrs []*Container,
	neighbours map[string]*Container,
) (rv map[string]*node.Node) {
	var (
		skip   bool
		notice bool
	)

	rv = make(map[string]*node.Node)

	for _, con := range cntrs {
		skip = !cfg.MatchName(con.Name)

		n := &node.Node{ID: con.ID, Name: con.Name, Image: con.Image, Volumes: []*node.Volume{}}

		if con.ConnectionsCount() == 0 && !notice {
			log.Printf("No connections for container: %s:%s, try run as root", con.ID, con.Name)

			notice = true
		}

		con.IterOutbounds(func(c *Connection) {
			rip := c.RemoteIP.String()
			rport := node.Port{Kind: c.Proto.String(), Value: int(c.RemotePort)}

			if lc, ok := neighbours[rip]; ok {
				if skip && cfg.MatchName(lc.Name) {
					skip = false
				}

				return
			}

			if cfg.OnlyLocal || skip {
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

		con.IterListeners(func(c *Connection) {
			port := node.Port{Kind: c.Proto.String(), Value: int(c.LocalPort)}

			n.Ports = append(n.Ports, port)
		})

		if con.Process != nil {
			n.Process = &node.Process{Cmd: con.Process.Cmd, Env: con.Process.Env}
		}

		if len(con.Volumes) > 0 {
			n.Volumes = make([]*node.Volume, len(con.Volumes))

			for i := 0; i < len(con.Volumes); i++ {
				cv := con.Volumes[i]
				n.Volumes[i] = &node.Volume{Type: cv.Type, Src: cv.Src, Dst: cv.Dst}
			}
		}

		rv[con.ID] = n
	}

	return rv
}

func directEdges(
	cfg *Config,
	cntrs []*Container,
	local map[string]*Container,
	nodes map[string]*node.Node,
) {
	for _, con := range cntrs {
		src, ok := nodes[con.ID]
		if !ok {
			continue
		}

		con.IterOutbounds(func(c *Connection) {
			port := node.Port{Kind: c.Proto.String(), Value: int(c.RemotePort)}
			key := c.RemoteIP.String()

			if ldst, ok := local[key]; ok {
				if cfg.NoLoops && con.ID == ldst.ID {
					return
				}

				if cfg.MatchName(ldst.Name) || cfg.MatchName(con.Name) {
					cfg.Builder.AddEdge(src.ID, ldst.ID, port)
				}

				return
			}

			if !cfg.MatchName(con.Name) {
				return
			}

			if rdst, ok := nodes[key]; ok {
				cfg.Builder.AddEdge(src.ID, rdst.ID, port)

				return
			}
		})
	}
}

func clusterEdges(
	cfg *Config,
	cntrs []*Container,
	local map[string]*Container,
	nodes map[string]*node.Node,
) {

}

func percentOf(a, b int) float64 {
	const hundred = 100.0

	return float64(a) / float64(b) * hundred
}
