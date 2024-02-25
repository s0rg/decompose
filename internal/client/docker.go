package client

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/graph"
)

const (
	stateRunning = "running"
	netstatCmd   = "netstat"
	netstatArg   = "-apn"
)

var ErrModeNone = errors.New("mode not set")

type (
	createClient func() (DockerClient, error)
	nsEnter      func(int, graph.NetProto, func(*graph.Connection)) error
)

type DockerClient interface {
	ContainerList(context.Context, container.ListOptions) ([]types.Container, error)
	ContainerInspect(context.Context, string) (types.ContainerJSON, error)
	ContainerExecCreate(context.Context, string, types.ExecConfig) (types.IDResponse, error)
	ContainerExecAttach(context.Context, string, types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerTop(ctx context.Context, containerID string, arguments []string) (container.ContainerTopOKBody, error)
	Close() error
}

type Docker struct {
	opt *options
	cli DockerClient
}

func NewDocker(opts ...Option) (rv *Docker, err error) {
	rv = &Docker{
		opt: &options{},
	}

	for _, op := range opts {
		op(rv.opt)
	}

	if rv.opt.Mode == None {
		return nil, fmt.Errorf("options: %w", ErrModeNone)
	}

	if rv.cli, err = rv.opt.Create(); err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return rv, nil
}

func (d *Docker) Mode() string {
	return d.opt.Mode.String()
}

func (d *Docker) Containers(
	ctx context.Context,
	proto graph.NetProto,
	detailed, deep bool,
	skipkeys []string,
	progress func(int, int),
) (rv []*graph.Container, err error) {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: true})
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
			ID:        doc.ID,
			Image:     doc.Image,
			Name:      strings.TrimLeft(doc.Names[0], "/"),
			Labels:    maps.Clone(doc.Labels),
			Endpoints: extractEndpoints(doc.NetworkSettings.Networks),
		}

		if detailed {
			info, err := d.cli.ContainerInspect(ctx, doc.ID)
			if err != nil {
				return nil, fmt.Errorf("inspect: %w", err)
			}

			con.Volumes = extractVolumesInfo(info.Mounts)
			con.Info = extractContainerInfo(&info, skeys)
		}

		if err := d.connections(ctx, doc.ID, proto, func(conn *graph.Connection) {
			if !deep && conn.LocalIP.IsLoopback() {
				return
			}

			con.AddConnection(conn)
		}); err != nil {
			return nil, fmt.Errorf("container: %s connections: %w", doc.ID, err)
		}

		con.SortConnections()

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

func (d *Docker) connections(
	ctx context.Context,
	cid string,
	proto graph.NetProto,
	cb func(*graph.Connection),
) (err error) {
	switch d.opt.Mode {
	case InContainer:
		err = d.connectionsContainer(ctx, cid, proto, func(r io.Reader) (err error) {
			if err = graph.ParseNetstat(r, cb); err != nil {
				return fmt.Errorf("parse: %w", err)
			}

			return nil
		})
	case LinuxNsenter:
		err = d.processesContainer(ctx, cid, func(pid int) (err error) {
			if err = d.opt.Nsenter(pid, proto, cb); err != nil {
				return fmt.Errorf("nsenter: %w", err)
			}

			return nil
		})
	}

	if err != nil {
		return fmt.Errorf("%s: %w", d.opt.Mode, err)
	}

	return nil
}

func (d *Docker) connectionsContainer(
	ctx context.Context,
	containerID string,
	proto graph.NetProto,
	parse func(io.Reader) error,
) (
	err error,
) {
	exe, err := d.cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		Privileged:   true,
		Cmd:          netstat(proto),
	})
	if err != nil {
		return fmt.Errorf("exec-create: %w", err)
	}

	resp, err := d.cli.ContainerExecAttach(ctx, exe.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		return fmt.Errorf("exec-attach: %w", err)
	}

	defer resp.Close()

	if err = parse(resp.Reader); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	return nil
}

func (d *Docker) processesContainer(
	ctx context.Context,
	cid string,
	fun func(int) error,
) (err error) {
	ps, err := d.cli.ContainerTop(ctx, cid, []string{"-o pid"})
	if err != nil {
		return fmt.Errorf("top: %w", err)
	}

	var pid int

	for _, p := range ps.Processes {
		if len(p) < len(ps.Titles) {
			continue
		}

		if pid, err = strconv.Atoi(p[0]); err != nil {
			continue
		}

		if err = fun(pid); err != nil {
			return fmt.Errorf("[pid: %d] %w", pid, err)
		}
	}

	return nil
}

func netstat(p graph.NetProto) []string {
	return []string{
		netstatCmd,
		netstatArg + p.Flag(),
	}
}

func extractContainerInfo(
	c *types.ContainerJSON,
	s set.Unordered[string],
) (rv *graph.ContainerInfo) {
	rv = &graph.ContainerInfo{
		Cmd: c.Config.Cmd,
	}

	if s.Len() == 0 {
		rv.Env = c.Config.Env

		return rv
	}

	const nparts = 2

	for _, env := range c.Config.Env {
		if key := strings.SplitN(env, "=", nparts)[0]; s.Has(key) {
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
	mounts []types.MountPoint,
) (rv []*graph.VolumeInfo) {
	rv = make([]*graph.VolumeInfo, len(mounts))

	for i, m := range mounts {
		rv[i] = &graph.VolumeInfo{
			Type: string(m.Type),
			Src:  m.Source,
			Dst:  m.Destination,
		}
	}

	slices.SortFunc(rv, func(a, b *graph.VolumeInfo) int {
		switch rv := cmp.Compare(a.Type, b.Type); rv {
		case 0:
			return cmp.Compare(a.Src, b.Src)
		default:
			return rv
		}
	})

	return rv
}
