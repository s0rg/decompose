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

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/set"
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

func checkState(state uint64) (listener, valid bool) {
	if state == tcpListen {
		return true, true
	}

	if state == tcpEstablished {
		return false, true
	}

	return false, false
}

func scanTCP(
	pfs procfs.FS,
	name string,
	onconn func(*graph.Connection),
) (err error) {
	tcp4, err := pfs.NetTCP()
	if err != nil {
		return fmt.Errorf("procfs/tcp4: %w", err)
	}

	for _, s := range tcp4 {
		listener, ok := checkState(s.St)
		if !ok {
			continue
		}

		onconn(&graph.Connection{
			Process: name,
			Inode:   s.Inode,
			SrcIP:   s.LocalAddr,
			DstIP:   s.RemAddr,
			SrcPort: int(s.LocalPort),
			DstPort: int(s.RemPort),
			Proto:   graph.TCP,
			Listen:  listener,
		})
	}

	tcp6, err := pfs.NetTCP6()
	if err != nil {
		return fmt.Errorf("procfs/tcp6: %w", err)
	}

	for _, s := range tcp6 {
		listener, ok := checkState(s.St)
		if !ok {
			continue
		}

		onconn(&graph.Connection{
			Process: name,
			Inode:   s.Inode,
			SrcIP:   s.LocalAddr,
			DstIP:   s.RemAddr,
			SrcPort: int(s.LocalPort),
			DstPort: int(s.RemPort),
			Proto:   graph.TCP,
			Listen:  listener,
		})
	}

	return nil
}

func scanUDP(
	pfs procfs.FS,
	name string,
	onconn func(*graph.Connection),
) (err error) {
	udp4, err := pfs.NetUDP()
	if err != nil {
		return fmt.Errorf("procfs/udp4: %w", err)
	}

	for _, s := range udp4 {
		onconn(&graph.Connection{
			Process: name,
			Inode:   s.Inode,
			SrcIP:   s.LocalAddr,
			DstIP:   s.RemAddr,
			SrcPort: int(s.LocalPort),
			DstPort: int(s.RemPort),
			Proto:   graph.UDP,
		})
	}

	udp6, err := pfs.NetUDP6()
	if err != nil {
		return fmt.Errorf("procfs/udp6: %w", err)
	}

	for _, s := range udp6 {
		onconn(&graph.Connection{
			Process: name,
			Inode:   s.Inode,
			SrcIP:   s.LocalAddr,
			DstIP:   s.RemAddr,
			SrcPort: int(s.LocalPort),
			DstPort: int(s.RemPort),
			Proto:   graph.UDP,
		})
	}

	return nil
}

func scanUNIX(
	pfs procfs.FS,
	name string,
	onconn func(*graph.Connection),
) (err error) {
	unix, err := pfs.NetUNIX()
	if err != nil {
		return fmt.Errorf("procfs/unix: %w", err)
	}

	for _, s := range unix.Rows {
		onconn(&graph.Connection{
			Process: name,
			Inode:   s.Inode,
			Path:    s.Path,
			Listen:  s.Flags != 0,
			Proto:   graph.UNIX,
		})
	}

	return nil
}

func processInfo(pid int) (
	name string,
	err error,
) {
	pfs, err := procfs.NewFS(procROOT)
	if err != nil {
		return "", fmt.Errorf("procfs: %w", err)
	}

	proc, err := pfs.Proc(pid)
	if err != nil {
		return "", fmt.Errorf("procfs/pid: %w", err)
	}

	name, err = proc.Executable()
	if err != nil {
		return "", fmt.Errorf("procfs/executable: %w", err)
	}

	return filepath.Base(name), nil
}

func Inodes(
	pid int,
	cb func(uint64),
) error {
	pfs, err := procfs.NewFS(procROOT)
	if err != nil {
		return fmt.Errorf("procfs: %w", err)
	}

	proc, err := pfs.Proc(pid)
	if err != nil {
		return fmt.Errorf("procfs/pid: %w", err)
	}

	fds, err := proc.FileDescriptorsInfo()
	if err != nil {
		return fmt.Errorf("procfs/descriptors: %w", err)
	}

	seen := make(set.Unordered[uint64])

	for _, f := range fds {
		ino, err := strconv.ParseUint(f.Ino, 10, 64)
		if err != nil {
			continue
		}

		if seen.Add(ino) {
			cb(ino)
		}
	}

	return nil
}

func Nsenter(
	pid int,
	proto graph.NetProto,
	onconn func(int, *graph.Connection),
) (
	err error,
) {
	name, err := processInfo(pid)
	if err != nil {
		return fmt.Errorf("procfs: %w", err)
	}

	connWithPid := func(c *graph.Connection) {
		onconn(pid, c)
	}

	fs, err := procfs.NewFS(filepath.Join(procROOT, strconv.Itoa(pid)))
	if err != nil {
		return fmt.Errorf("procfs/net: %w", err)
	}

	if proto == graph.ALL || proto == graph.TCP {
		if err = scanTCP(fs, name, connWithPid); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	if proto == graph.ALL || proto == graph.UDP {
		if err = scanUDP(fs, name, connWithPid); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	if proto == graph.ALL || proto == graph.UNIX {
		if err = scanUNIX(fs, name, connWithPid); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	return nil
}
