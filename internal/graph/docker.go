//go:build !test

package graph

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	stateRunning = "running"
	nsenterCmd   = "nsenter"
	pingTimeout  = time.Second
)

type DockerClient struct {
	cli *client.Client
	cmd string
}

func NewDockerClient() (rv *DockerClient, err error) {
	rv = &DockerClient{}

	if rv.cmd, err = exec.LookPath(nsenterCmd); err != nil {
		return nil, fmt.Errorf("looking for %s: %w", nsenterCmd, err)
	}

	if rv.cli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	); err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if _, err = rv.cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return rv, nil
}

func (d *DockerClient) Containers(
	ctx context.Context,
	kind NetProto,
	progress func(int, int),
) (rv []*Container, err error) {
	containers, err := d.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	rv = make([]*Container, 0, len(containers))

	for i := 0; i < len(containers); i++ {
		doc := &containers[i]

		if doc.State != stateRunning {
			continue
		}

		ic := &Container{
			ID:    doc.ID,
			Image: doc.Image,
			Name:  strings.TrimLeft(doc.Names[0], "/"),
		}

		conns, err := d.conections(ctx, doc.ID, kind)
		if err != nil {
			return nil, fmt.Errorf("container: %s connections: %w", doc.ID, err)
		}

		ic.SetConnections(conns)

		ic.Endpoints = make(map[string]string)

		for name, n := range doc.NetworkSettings.Networks {
			if n.EndpointID == "" {
				continue
			}

			ic.Endpoints[n.IPAddress] = name
		}

		rv = append(rv, ic)

		progress(i, len(containers))
	}

	progress(len(containers), len(containers))

	return slices.Clip(rv), nil
}

func (d *DockerClient) Close() (err error) {
	if err = d.cli.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return nil
}

func (d *DockerClient) conections(
	ctx context.Context,
	containerID string,
	kind NetProto,
) (
	rv []*Connection,
	err error,
) {
	info, err := d.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("inspect: %w", err)
	}

	arg := []string{"-t", strconv.Itoa(info.State.Pid), "-n", "netstat", "-an" + kind.Flag()}
	cmd := exec.CommandContext(ctx, d.cmd, arg...)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("pipe: %w", err)
	}

	defer pipe.Close()

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}

	if rv, err = ParseNetstat(pipe); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return rv, nil
}
