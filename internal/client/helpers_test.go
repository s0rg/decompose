package client_test

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

func voidProgress(_, _ int) {}

type clientMock struct {
	Err          error
	OnList       func() []types.Container
	OnInspect    func() types.ContainerJSON
	OnExecCreate func() types.IDResponse
	OnExecAttach func() types.HijackedResponse
}

func (cm *clientMock) ContainerList(
	_ context.Context,
	_ container.ListOptions,
) (rv []types.Container, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnList(), nil
}

func (cm *clientMock) ContainerInspect(
	_ context.Context,
	_ string,
) (rv types.ContainerJSON, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnInspect(), nil
}

func (cm *clientMock) ContainerExecCreate(
	_ context.Context,
	_ string,
	_ types.ExecConfig,
) (rv types.IDResponse, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnExecCreate(), nil
}

func (cm *clientMock) ContainerExecAttach(
	_ context.Context,
	_ string,
	_ types.ExecStartCheck,
) (rv types.HijackedResponse, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnExecAttach(), nil
}

func (cm *clientMock) Ping(_ context.Context) (rv types.Ping, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return rv, nil
}

func (cm *clientMock) Close() (err error) {
	if cm.Err != nil {
		err = cm.Err
	}

	return
}

type connMock struct {
	Err error
}

func (cnm *connMock) Read(_ []byte) (n int, err error) {
	if cnm.Err != nil {
		return 0, cnm.Err
	}

	return 0, io.EOF
}

func (cnm *connMock) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (cnm *connMock) Close() error {
	return cnm.Err
}

func (cnm *connMock) LocalAddr() (rv net.Addr)           { return }
func (cnm *connMock) RemoteAddr() (rv net.Addr)          { return }
func (cnm *connMock) SetDeadline(_ time.Time) error      { return nil }
func (cnm *connMock) SetReadDeadline(_ time.Time) error  { return nil }
func (cnm *connMock) SetWriteDeadline(_ time.Time) error { return nil }
