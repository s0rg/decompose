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
	"github.com/otterize/go-procnet/procnet"

	"github.com/s0rg/decompose/internal/graph"
)

const (
	pingTimeout = time.Second
	procENV     = "IN_DOCKER_PROC_ROOT"
	procDefault = "/proc"
	procNET     = "net"
	procTCP4    = "tcp"
	procTCP6    = "tcp6"
	procUDP4    = "udp"
	procUDP6    = "udp6"
)

var procROOT = procDefault

type sockCB func(procnet.SockTabEntry)

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

func scanTCP(pid string, onconn connCB) (err error) {
	onsock := func(s procnet.SockTabEntry) {
		switch s.State {
		case procnet.Listen, procnet.Established:
			onconn(
				s.LocalAddr.IP,
				s.RemoteAddr.IP,
				s.LocalAddr.Port,
				s.RemoteAddr.Port,
				procTCP4,
			)
		default:
		}
	}

	pathTCP4 := filepath.Join(procROOT, pid, procNET, procTCP4)
	if err = scan(pathTCP4, onsock); err != nil {
		return fmt.Errorf("tcp4: %w", err)
	}

	pathTCP6 := filepath.Join(procROOT, pid, procNET, procTCP6)
	if err = scan(pathTCP6, onsock); err != nil {
		return fmt.Errorf("tcp6: %w", err)
	}

	return nil
}

func scanUDP(pid string, onconn connCB) (err error) {
	onsock := func(s procnet.SockTabEntry) {
		onconn(
			s.LocalAddr.IP,
			s.RemoteAddr.IP,
			s.LocalAddr.Port,
			s.RemoteAddr.Port,
			procUDP4,
		)
	}

	pathUDP4 := filepath.Join(procROOT, pid, procNET, procUDP4)
	if err = scan(pathUDP4, onsock); err != nil {
		return fmt.Errorf("udp4: %w", err)
	}

	pathUDP6 := filepath.Join(procROOT, pid, procNET, procUDP6)
	if err = scan(pathUDP6, onsock); err != nil {
		return fmt.Errorf("udp6: %w", err)
	}

	return nil
}

func scan(path string, onsock sockCB) (err error) {
	socks, err := procnet.SocksFromPath(path)
	if err != nil {
		return fmt.Errorf("path: %w", err)
	}

	for _, s := range socks {
		onsock(s)
	}

	return nil
}

func Nsenter(
	pid int,
	proto graph.NetProto,
	onconn connCB,
) (
	err error,
) {
	spid := strconv.Itoa(pid)

	if proto == graph.ALL || proto == graph.TCP {
		if err = scanTCP(spid, onconn); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	if proto == graph.ALL || proto == graph.UDP {
		if err = scanUDP(spid, onconn); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	return nil
}
