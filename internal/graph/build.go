package graph

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/s0rg/decompose/internal/node"
)

const (
	minItems  = 2
	minReport = 10

	ProcessRemote  = "[remote]"
	ProcessUnknown = "[unknown]"
)

var ErrNotEnough = errors.New("not enough items")

type ContainerClient interface {
	Containers(context.Context, NetProto, bool, []string, func(int, int)) ([]*Container, error)
}

type Builder interface {
	AddNode(*node.Node) error
	AddEdge(*node.Edge)
}

type NamedWriter interface {
	Name() string
	Write(io.Writer) error
}

type NamedBuilderWriter interface {
	Builder
	NamedWriter
}

type Enricher interface {
	Enrich(*node.Node)
}

func Build(
	cfg *Config,
	cli ContainerClient,
) error {
	log.Println("Gathering containers info")

	containers, err := cli.Containers(
		context.Background(),
		cfg.Proto,
		cfg.Deep,
		cfg.SkipEnv,
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
		cfg.Meta.Enrich(node)

		if err = cfg.Builder.AddNode(node); err != nil {
			return fmt.Errorf("node '%s': %w", node.Name, err)
		}
	}

	log.Println("Building edges")

	buildEdges(cfg, containers, neighbours, nodes)

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
	local map[string]*Container,
) (rv map[string]*node.Node) {
	var (
		skip   bool
		notice bool
	)

	rv = make(map[string]*node.Node)

	for _, con := range cntrs {
		if con.ConnectionsCount() == 0 && !notice {
			log.Printf("No connections for container: %s:%s, try run as root", con.ID, con.Name)

			notice = true
		}

		skip = !cfg.MatchName(con.Name)

		con.IterOutbounds(func(c *Connection) {
			if c.Proto == UNIX || c.DstIP.IsLoopback() {
				return
			}

			rip := c.DstIP.String()

			if lc, ok := local[rip]; ok { // destination known
				if skip && cfg.MatchName(lc.Name) {
					skip = false
				}

				return
			}

			if cfg.OnlyLocal || skip {
				return
			}

			// destination is remote host, add it
			rem, ok := rv[rip]
			if !ok {
				rem = node.External(rip)
				rv[rip] = rem
			}

			rem.Ports.Add(ProcessRemote, &node.Port{
				Kind:  c.Proto.String(),
				Value: strconv.Itoa(c.DstPort),
			})
		})

		if !skip {
			rv[con.ID] = con.ToNode()
		}
	}

	return rv
}

func buildEdges(
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
			var (
				port = &node.Port{
					Kind: c.Proto.String(),
				}
				dstID string
				key   string
				ok    bool
			)

			if c.Proto == UNIX {
				port.Value = c.Path
				dstID, ok = c.DstID, true
			} else {
				key = c.DstIP.String()
				port.Value = strconv.Itoa(c.DstPort)

				var ldst *Container

				if c.DstIP.IsLoopback() {
					ldst, ok = con, true
				} else {
					ldst, ok = local[key]
				}

				if ok {
					dstID = ldst.ID
				}
			}

			if ok {
				if cfg.NoLoops && con.ID == dstID {
					return
				}

				dst, found := nodes[dstID]
				if !found {
					return
				}

				dname, found := dst.Ports.Get(port)
				if !found {
					dname = ProcessUnknown
				}

				cfg.Builder.AddEdge(&node.Edge{
					SrcID:   src.ID,
					SrcName: c.Process,
					DstID:   dstID,
					DstName: dname,
					Port:    port,
				})

				return
			}

			if !cfg.MatchName(con.Name) {
				return
			}

			if rdst, ok := nodes[key]; ok {
				cfg.Builder.AddEdge(&node.Edge{
					SrcID:   src.ID,
					SrcName: c.Process,
					DstID:   rdst.ID,
					DstName: ProcessRemote,
					Port:    port,
				})

				return
			}
		})
	}
}

func percentOf(a, b int) float64 {
	const hundred = 100.0

	return float64(a) / float64(b) * hundred
}
