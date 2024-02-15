//go:build !test

package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/prometheus/procfs"
	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/graph"
)

const (
	pingTimeout = time.Second
	procENV     = "IN_DOCKER_PROC_ROOT"
	procDefault = "/proc"

	// from net/tcp_states.h.
	tcpEstablished = uint64(1)
	tcpListen      = uint64(10)
)

var procROOT = procDefault

func Default() (rv DockerClient, err error) {
	var dc *client.Client

	dc, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("docker: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if _, err = dc.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	if root := os.Getenv(procENV); root != "" {
		procROOT = filepath.Join(root, procDefault)
	}

	return dc, nil
}

func isValidState(state uint64) (ok bool) {
	switch state {
	case tcpEstablished, tcpListen:
		ok = true
	default:
	}

	return ok
}

func scanTCP(
	pfs procfs.FS,
	name string,
	inodes set.Unordered[uint64],
	onconn func(*graph.Connection),
) (err error) {
	tcp4, err := pfs.NetTCP()
	if err != nil {
		return fmt.Errorf("procfs/tcp4: %w", err)
	}

	for _, s := range tcp4 {
		if !isValidState(s.St) {
			continue
		}

		if !inodes.Has(s.Inode) {
			continue
		}

		onconn(&graph.Connection{
			LocalIP:    s.LocalAddr,
			RemoteIP:   s.RemAddr,
			LocalPort:  uint16(s.LocalPort),
			RemotePort: uint16(s.RemPort),
			Proto:      graph.TCP,
			Process:    name,
		})
	}

	tcp6, err := pfs.NetTCP6()
	if err != nil {
		return fmt.Errorf("procfs/tcp6: %w", err)
	}

	for _, s := range tcp6 {
		if !isValidState(s.St) {
			continue
		}

		if !inodes.Has(s.Inode) {
			continue
		}

		onconn(&graph.Connection{
			LocalIP:    s.LocalAddr,
			RemoteIP:   s.RemAddr,
			LocalPort:  uint16(s.LocalPort),
			RemotePort: uint16(s.RemPort),
			Proto:      graph.TCP,
			Process:    name,
		})
	}

	return nil
}

func scanUDP(
	pfs procfs.FS,
	name string,
	inodes set.Unordered[uint64],
	onconn func(*graph.Connection),
) (err error) {
	udp4, err := pfs.NetUDP()
	if err != nil {
		return fmt.Errorf("procfs/udp4: %w", err)
	}

	for _, s := range udp4 {
		if !inodes.Has(s.Inode) {
			continue
		}

		onconn(&graph.Connection{
			LocalIP:    s.LocalAddr,
			RemoteIP:   s.RemAddr,
			LocalPort:  uint16(s.LocalPort),
			RemotePort: uint16(s.RemPort),
			Proto:      graph.UDP,
			Process:    name,
		})
	}

	udp6, err := pfs.NetUDP6()
	if err != nil {
		return fmt.Errorf("procfs/udp6: %w", err)
	}

	for _, s := range udp6 {
		if !inodes.Has(s.Inode) {
			continue
		}

		onconn(&graph.Connection{
			LocalIP:    s.LocalAddr,
			RemoteIP:   s.RemAddr,
			LocalPort:  uint16(s.LocalPort),
			RemotePort: uint16(s.RemPort),
			Proto:      graph.UDP,
			Process:    name,
		})
	}

	return nil
}

func processInfo(pid int) (
	name string,
	inodes set.Unordered[uint64],
	err error,
) {
	pfs, err := procfs.NewFS(procROOT)
	if err != nil {
		return "", nil, fmt.Errorf("procfs: %w", err)
	}

	proc, err := pfs.Proc(pid)
	if err != nil {
		return "", nil, fmt.Errorf("procfs/pid: %w", err)
	}

	name, err = proc.Executable()
	if err != nil {
		return "", nil, fmt.Errorf("procfs/executable: %w", err)
	}

	fds, err := proc.FileDescriptorsInfo()
	if err != nil {
		return "", nil, fmt.Errorf("procfs/descriptors: %w", err)
	}

	inodes = make(set.Unordered[uint64])

	for _, f := range fds {
		ino, err := strconv.ParseUint(f.Ino, 10, 64)
		if err != nil {
			continue
		}

		inodes.Add(ino)
	}

	return filepath.Base(name), inodes, nil
}

func Nsenter(
	pid int,
	proto graph.NetProto,
	onconn func(*graph.Connection),
) (
	err error,
) {
	name, inodes, err := processInfo(pid)
	if err != nil {
		return fmt.Errorf("procfs: %w", err)
	}

	fs, err := procfs.NewFS(filepath.Join(procROOT, strconv.Itoa(pid)))
	if err != nil {
		return fmt.Errorf("procfs/net: %w", err)
	}

	if proto == graph.ALL || proto == graph.TCP {
		if err = scanTCP(fs, name, inodes, onconn); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	if proto == graph.ALL || proto == graph.UDP {
		if err = scanUDP(fs, name, inodes, onconn); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	return nil
}
