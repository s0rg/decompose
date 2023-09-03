//go:build !test

package client

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
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/graph"
)

const (
	linuxOS      = "linux"
	stateRunning = "running"
	nsenterCmd   = "nsenter"
	pingTimeout  = time.Second
)

type connExtactor func(context.Context, int, string, graph.NetProto) ([]*graph.Connection, error)

type Docker struct {
	cli         *client.Client
	connections connExtactor
	cmd         string
	kind        string
}

func NewDocker() (rv *Docker, err error) {
	rv = &Docker{}

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

func (d *Docker) init() (err error) {
	if runtime.GOOS == linuxOS && os.Geteuid() == 0 {
		if d.cmd, err = exec.LookPath(nsenterCmd); err != nil {
			return fmt.Errorf("looking for %s: %w", nsenterCmd, err)
		}

		d.kind = "netns"
		d.connections = d.conectionsNetns

		return nil
	}

	// non-linux or non-root
	d.kind = "in-container"
	d.connections = d.conectionsContainer

	return nil
}

func (d *Docker) Kind() string {
	return d.kind
}

func (d *Docker) Containers(
	ctx context.Context,
	proto graph.NetProto,
	detailed bool,
	skipkeys []string,
	progress func(int, int),
) (rv []*graph.Container, err error) {
	containers, err := d.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	skeys := make(set.Unordered[string])

	for _, key := range skipkeys {
		skeys.Add(strings.ToUpper(key))
	}

	rv = make([]*graph.Container, 0, len(containers))

	for i := 0; i < len(containers); i++ {
		doc := &containers[i]

		if doc.State != stateRunning {
			continue
		}

		con := &graph.Container{
			ID:    doc.ID,
			Image: doc.Image,
			Name:  strings.TrimLeft(doc.Names[0], "/"),
		}

		var pid int

		if detailed {
			info, err := d.cli.ContainerInspect(ctx, doc.ID)
			if err != nil {
				return nil, fmt.Errorf("inspect: %w", err)
			}

			con.Volumes = extractVolumesInfo(doc)
			con.Process = extractProcessInfo(&info, skeys)
			pid = info.State.Pid
		}

		conns, err := d.connections(ctx, pid, doc.ID, proto)
		if err != nil {
			return nil, fmt.Errorf("container: %s connections: %w", doc.ID, err)
		}

		con.SetConnections(conns)

		con.Endpoints = extractEndpoints(doc.NetworkSettings.Networks)

		rv = append(rv, con)

		progress(i, len(containers))
	}

	progress(len(containers), len(containers))

	return slices.Clip(rv), nil
}

func (d *Docker) Close() (err error) {
	if err = d.cli.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return nil
}

func (d *Docker) conectionsNetns(
	ctx context.Context,
	pid int,
	containerID string,
	proto graph.NetProto,
) (
	rv []*graph.Connection,
	err error,
) {
	if pid == 0 {
		info, ierr := d.cli.ContainerInspect(ctx, containerID)
		if ierr != nil {
			return nil, fmt.Errorf("inspect: %w", ierr)
		}

		pid = info.State.Pid
	}

	arg := append([]string{"-t", strconv.Itoa(pid), "-n"}, netstatCmd(proto)...)
	cmd := exec.CommandContext(ctx, d.cmd, arg...)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("pipe: %w", err)
	}

	defer pipe.Close()

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}

	if rv, err = graph.ParseNetstat(pipe); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return rv, nil
}

func (d *Docker) conectionsContainer(
	ctx context.Context,
	_ int,
	containerID string,
	proto graph.NetProto,
) (
	rv []*graph.Connection,
	err error,
) {
	exe, err := d.cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		Cmd:          netstatCmd(proto),
	})
	if err != nil {
		return nil, fmt.Errorf("exec-create: %w", err)
	}

	resp, err := d.cli.ContainerExecAttach(ctx, exe.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		return nil, fmt.Errorf("exec-attach: %w", err)
	}

	defer resp.Close()

	if rv, err = graph.ParseNetstat(resp.Reader); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return rv, nil
}

func netstatCmd(p graph.NetProto) []string {
	return []string{"netstat", "-an" + p.Flag()}
}

func extractProcessInfo(
	c *types.ContainerJSON,
	s set.Unordered[string],
) (rv *graph.ProcessInfo) {
	rv = &graph.ProcessInfo{Cmd: c.Config.Cmd}

	if s.Len() == 0 {
		rv.Env = c.Config.Env

		return rv
	}

	const nparts = 2

	for _, env := range c.Config.Env {
		key := strings.SplitN(env, "=", nparts)[0]
		if s.Has(key) {
			continue
		}

		rv.Env = append(rv.Env, env)
	}

	return rv
}

func extractEndpoints(
	nets map[string]*network.EndpointSettings,
) (rv map[string]string) {
	rv = make(map[string]string)

	for name, n := range nets {
		if n.EndpointID == "" {
			continue
		}

		rv[n.IPAddress] = name
	}

	return rv
}

func extractVolumesInfo(
	c *types.Container,
) (rv []*graph.VolumeInfo) {
	rv = make([]*graph.VolumeInfo, len(c.Mounts))

	for i, m := range c.Mounts {
		rv[i] = &graph.VolumeInfo{
			Type: string(m.Type),
			Src:  m.Source,
			Dst:  m.Destination,
		}
	}

	return rv
}
