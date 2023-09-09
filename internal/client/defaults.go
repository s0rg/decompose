//go:build !test

package client

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/s0rg/decompose/internal/graph"
)

func Default() (rv DockerClient, err error) {
	rv, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("docker: %w", err)
	}

	return rv, nil
}

func Nsenter(
	ctx context.Context,
	pid int,
	proto graph.NetProto,
	parse func(io.Reader) error,
) (
	err error,
) {
	arg := append([]string{"-t", strconv.Itoa(pid), "-n"}, netstat(proto)...)
	cmd := exec.CommandContext(ctx, nsenterCmd, arg...)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe: %w", err)
	}

	defer pipe.Close()

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if err = parse(pipe); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	return nil
}
