//go:build !test

package graph

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	linuxOS      = "linux"
	stateRunning = "running"
	nsenterCmd   = "nsenter"
	pingTimeout  = time.Second
)

type DockerClient struct {
	cli         *client.Client
	connections func(context.Context, string, NetProto) ([]*Connection, error)
	cmd         string
	kind        string
}

func NewDockerClient() (rv *DockerClient, err error) {
	rv = &DockerClient{}

	if err = rv.init(); err != nil {
		return nil, fmt.Errorf("init: %w", err)
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

func (d *DockerClient) init() (err error) {
	d.kind = runtime.GOOS

	if runtime.GOOS == linuxOS && os.Geteuid() == 0 {
		if d.cmd, err = exec.LookPath(nsenterCmd); err != nil {
			return fmt.Errorf("looking for %s: %w", nsenterCmd, err)
		}

		d.kind += "/netns"
		d.connections = d.conectionsNetns

		return nil
	}

	// non-linux or non-root
	d.kind += "/container"
	d.connections = d.conectionsContainer

	return nil
}

func (d *DockerClient) Kind() string {
	return d.kind
}

func (d *DockerClient) Containers(
	ctx context.Context,
	proto NetProto,
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

		conns, err := d.connections(ctx, doc.ID, proto)
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

func (d *DockerClient) conectionsNetns(
	ctx context.Context,
	containerID string,
	proto NetProto,
) (
	rv []*Connection,
	err error,
) {
	info, err := d.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("inspect: %w", err)
	}

	arg := append([]string{"-t", strconv.Itoa(info.State.Pid), "-n"}, netstatCmd(proto)...)
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

func (d *DockerClient) conectionsContainer(
	ctx context.Context,
	containerID string,
	proto NetProto,
) (
	rv []*Connection,
	err error,
) {
	exe, err := d.cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		Cmd:          netstatCmd(proto),
	})
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	resp, err := d.cli.ContainerExecAttach(ctx, exe.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		return nil, fmt.Errorf("attach: %w", err)
	}

	defer resp.Close()

	if rv, err = ParseNetstat(resp.Reader); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return rv, nil
}

func netstatCmd(p NetProto) []string {
	return []string{"netstat", "-an" + p.Flag()}
}
