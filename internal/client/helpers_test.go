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
	Err            error
	OnList         func() []container.Summary
	OnInspect      func() container.InspectResponse
	OnExecCreate   func() container.ExecCreateResponse
	OnExecAttach   func() types.HijackedResponse
	OnContainerTop func() container.TopResponse
}

func (cm *clientMock) ContainerTop(
	_ context.Context,
	_ string,
	_ []string,
) (rv container.TopResponse, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnContainerTop(), nil
}

func (cm *clientMock) ContainerList(
	_ context.Context,
	_ container.ListOptions,
) (rv []container.Summary, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnList(), nil
}

func (cm *clientMock) ContainerInspect(
	_ context.Context,
	_ string,
) (rv container.InspectResponse, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnInspect(), nil
}

func (cm *clientMock) ContainerExecCreate(
	_ context.Context,
	_ string,
	_ container.ExecOptions,
) (rv container.ExecCreateResponse, err error) {
	if cm.Err != nil {
		err = cm.Err

		return
	}

	return cm.OnExecCreate(), nil
}

func (cm *clientMock) ContainerExecAttach(
	_ context.Context,
	_ string,
	_ container.ExecStartOptions,
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
