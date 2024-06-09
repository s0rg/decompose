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
)

var ErrModeNone = errors.New("mode not set")

type (
	createClient func() (DockerClient, error)
	nsEnter      func(int, graph.NetProto, func(int, *graph.Connection)) error
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
	unix, deep bool,
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

	var inodes *InodesMap

	if unix {
		if inodes, err = d.collectInodes(ctx, containers); err != nil {
			return nil, fmt.Errorf("inodes: %w", err)
		}
	}

	cmap := make(map[string]*graph.Container)

	for i := 0; i < len(containers); i++ {
		c := &containers[i]

		if c.State != stateRunning {
			continue
		}

		con, cerr := d.extractInfo(ctx, c, proto, unix, deep, skeys, inodes)
		if cerr != nil {
			return nil, fmt.Errorf("container %s: %w", c.ID, cerr)
		}

		cmap[c.ID] = con

		rv = append(rv, con)

		progress(i, len(containers))
	}

	if unix {
		inodes.ResolveUnknown(func(srcCID, dstCID, srcName, _, path string) {
			dst, ok := cmap[dstCID]
			if !ok {
				return
			}

			dst.AddConnection(&graph.Connection{
				Proto:   graph.UNIX,
				Process: srcName,
				Path:    path,
				DstID:   srcCID,
			})
		})
	}

	progress(len(containers), len(containers))

	return slices.Clip(rv), nil
}

func (d *Docker) collectInodes(
	ctx context.Context,
	containers []types.Container,
) (
	inodes *InodesMap,
	err error,
) {
	inodes = &InodesMap{}

	for i := 0; i < len(containers); i++ {
		c := &containers[i]

		if c.State != stateRunning {
			continue
		}

		err = d.processesContainer(ctx, c.ID, func(pid int, name string) (err error) {
			inodes.AddProcess(c.ID, pid, name)

			return Inodes(pid, func(inode uint64) {
				inodes.AddInode(c.ID, pid, inode)
			})
		})
		if err != nil {
			return nil, fmt.Errorf("inodes %s: %w", c.ID, err)
		}
	}

	return inodes, nil
}

func (d *Docker) extractInfo(
	ctx context.Context,
	c *types.Container,
	proto graph.NetProto,
	unix, deep bool,
	skeys set.Unordered[string],
	inodes *InodesMap,
) (rv *graph.Container, err error) {
	rv = &graph.Container{
		ID:        c.ID,
		Image:     c.Image,
		Name:      strings.TrimLeft(c.Names[0], "/"),
		Labels:    maps.Clone(c.Labels),
		Endpoints: extractEndpoints(c.NetworkSettings.Networks),
	}

	info, err := d.cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("inspect: %w", err)
	}

	rv.Volumes = extractVolumesInfo(info.Mounts)
	rv.Info = extractContainerInfo(&info, skeys)

	if err := d.connections(ctx, c.ID, proto, func(pid int, conn *graph.Connection) {
		if !deep && conn.IsLocal() {
			return
		}

		if !unix && conn.Proto == graph.UNIX {
			return
		}

		if conn.Proto == graph.UNIX {
			if !inodes.Has(c.ID, pid, conn.Inode) {
				inodes.MarkUnknown(c.ID, pid, conn.Inode)

				return
			}

			if conn.Listen {
				inodes.MarkListener(c.ID, pid, conn.Path)
			}
		}

		rv.AddConnection(conn)
	}); err != nil {
		return nil, fmt.Errorf("connections: %w", err)
	}

	rv.SortConnections()

	return rv, nil
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
	cb func(int, *graph.Connection),
) (err error) {
	switch d.opt.Mode {
	case InContainer:
		err = d.connectionsContainer(ctx, cid, proto, func(r io.Reader) (err error) {
			if err = graph.ParseNetstat(r, func(c *graph.Connection) {
				cb(1, c)
			}); err != nil {
				return fmt.Errorf("parse: %w", err)
			}

			return nil
		})
	case LinuxNsenter:
		err = d.processesContainer(ctx, cid, func(pid int, _ string) (err error) {
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
		Cmd:          graph.NetstatCMD(proto),
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
	fun func(int, string) error,
) (err error) {
	ps, err := d.cli.ContainerTop(ctx, cid, []string{"-o pid,cmd"})
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

		cmd := strings.Fields(p[1])

		if err = fun(pid, cmd[0]); err != nil {
			return fmt.Errorf("[pid: %d] %w", pid, err)
		}
	}

	return nil
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
