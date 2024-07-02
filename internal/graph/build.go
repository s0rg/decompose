package graph

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

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
	log.Println("Gathering containers info, please be patient...")

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

	state := newBuilderState(cfg, containers)

	log.Println("Building nodes...")

	nodes, err := state.BuildNodes()
	if err != nil {
		return fmt.Errorf("build nodes: %w", err)
	}

	log.Printf("Processing %d nodes", nodes)

	if nodes < minItems {
		return fmt.Errorf("%w: nodes", ErrNotEnough)
	}

	log.Println("Building edges...")

	log.Printf("Found %d edges", state.BuildEdges())

	return nil
}

func percentOf(a, b int) float64 {
	const hundred = 100.0

	return float64(a) / float64(b) * hundred
}
