package client_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"github.com/s0rg/decompose/internal/client"
	"github.com/s0rg/decompose/internal/graph"
)

func TestDockerClientCreateModeError(t *testing.T) {
	t.Parallel()

	_, err := client.NewDocker()
	if err == nil {
		t.Fail()
	}
}

func TestDockerClientCreateError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test-error")

	_, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return nil, testErr
		}),
		client.WithMode(client.InContainer),
	)
	if err == nil || !errors.Is(err, testErr) {
		t.Fail()
	}
}

func TestDockerClientContainersError(t *testing.T) {
	t.Parallel()

	cm := &clientMock{}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal(err)
	}

	cm.Err = errors.New("test-error")

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err == nil || !errors.Is(err, cm.Err) {
		t.Fail()
	}
}

func TestDockerClientContainersEmpty(t *testing.T) {
	t.Parallel()

	cm := &clientMock{
		OnList: func() (rv []types.Container) {
			return rv
		},
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	rv, err := cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}

	if len(rv) != 0 {
		t.Fail()
	}
}

func TestDockerClientContainersSingleExited(t *testing.T) {
	t.Parallel()

	cm := &clientMock{
		OnList: func() (rv []types.Container) {
			return []types.Container{
				{
					State: "exited",
				},
			}
		},
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	rv, err := cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}

	if len(rv) != 0 {
		t.Fail()
	}
}

func TestDockerClientContainersExecCreateError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test-err")

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		cm.Err = testErr

		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil && !errors.Is(err, cm.Err) {
		t.Fail()
	}
}

func TestDockerClientContainersInspectError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test-err")

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		cm.Err = testErr

		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil && !errors.Is(err, cm.Err) {
		t.Fail()
	}
}

func TestDockerClientContainersExecAttachError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test-err")

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cm.OnExecCreate = func() (rv types.IDResponse) {
		cm.Err = testErr

		return
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil && !errors.Is(err, cm.Err) {
		t.Fail()
	}
}

func TestDockerClientContainersParseError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test-err")

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cm.OnExecCreate = func() (rv types.IDResponse) {
		return
	}

	cm.OnExecAttach = func() (rv types.HijackedResponse) {
		rv.Conn = &connMock{}
		rv.Reader = bufio.NewReader(&connMock{Err: testErr})

		return
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil && !errors.Is(err, testErr) {
		t.Fail()
	}
}

func TestDockerClientContainersCloseError(t *testing.T) {
	t.Parallel()

	cm := &clientMock{}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	cm.Err = errors.New("test-err")

	if err = cli.Close(); !errors.Is(err, cm.Err) {
		t.Fail()
	}
}

func TestDockerClientContainersSingle(t *testing.T) {
	t.Parallel()

	cm := &clientMock{
		OnList: func() (rv []types.Container) {
			return []types.Container{
				{
					ID:    "1",
					Names: []string{"test"},
					Image: "test-image",
					State: "running",
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"test-net": {
								EndpointID: "1",
								IPAddress:  "1.1.1.1",
							},
							"empty-id": {
								IPAddress: "1.1.1.2",
							},
						},
					},
				},
			}
		},
		OnExecCreate: func() (rv types.IDResponse) {
			return
		},
		OnExecAttach: func() (rv types.HijackedResponse) {
			rv.Conn = &connMock{}
			rv.Reader = bufio.NewReader(bytes.NewBufferString(""))

			return
		},
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	rv, err := cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}

	cli.Close()

	if len(rv) != 1 {
		t.Fail()
	}

	if rv[0].Name != "test" {
		t.Fail()
	}

	if len(rv[0].Endpoints) != 1 {
		t.Fail()
	}
}

func TestDockerClientContainersSingleFull(t *testing.T) {
	t.Parallel()

	cm := &clientMock{
		OnList: func() (rv []types.Container) {
			return []types.Container{
				{
					ID:    "1",
					Names: []string{"test"},
					Image: "test-image",
					State: "running",
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"test-net": {
								EndpointID: "1",
								IPAddress:  "1.1.1.1",
							},
							"empty-id": {
								IPAddress: "1.1.1.2",
							},
						},
					},
				},
			}
		},
		OnInspect: func() (rv types.ContainerJSON) {
			rv.ContainerJSONBase = &types.ContainerJSONBase{}
			rv.State = &types.ContainerState{Pid: 1}
			rv.Config = &container.Config{
				Cmd: []string{"foo"},
				Env: []string{"BAR=1"},
			}
			rv.Mounts = []types.MountPoint{
				{
					Type:        "bind",
					Source:      "src",
					Destination: "dst",
				},
				{
					Type:        "bind",
					Source:      "src1",
					Destination: "dst1",
				},
				{
					Type:        "mount",
					Source:      "src2",
					Destination: "dst2",
				},
			}

			return rv
		},
		OnExecCreate: func() (rv types.IDResponse) {
			return
		},
		OnExecAttach: func() (rv types.HijackedResponse) {
			rv.Conn = &connMock{}
			rv.Reader = bufio.NewReader(bytes.NewBufferString(""))

			return
		},
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	rv, err := cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}

	if len(rv) != 1 {
		t.Fail()
	}

	if rv[0].Info == nil {
		t.Fail()
	}

	if len(rv[0].Volumes) != 3 {
		t.Fail()
	}
}

func TestDockerClientContainersSingleFullSkipEnv(t *testing.T) {
	t.Parallel()

	cm := &clientMock{
		OnList: func() (rv []types.Container) {
			return []types.Container{
				{
					ID:    "1",
					Names: []string{"test"},
					Image: "test-image",
					State: "running",
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"test-net": {
								EndpointID: "1",
								IPAddress:  "1.1.1.1",
							},
							"empty-id": {
								IPAddress: "1.1.1.2",
							},
						},
					},
				},
			}
		},
		OnInspect: func() (rv types.ContainerJSON) {
			rv.ContainerJSONBase = &types.ContainerJSONBase{}
			rv.State = &types.ContainerState{Pid: 1}
			rv.Config = &container.Config{
				Cmd: []string{"foo"},
				Env: []string{"BAR=1", "BAZ=2"},
			}
			rv.Mounts = []types.MountPoint{}

			return rv
		},
		OnExecCreate: func() (rv types.IDResponse) {
			return
		},
		OnExecAttach: func() (rv types.HijackedResponse) {
			rv.Conn = &connMock{}
			rv.Reader = bufio.NewReader(bytes.NewBufferString(""))

			return
		},
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	rv, err := cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		[]string{"BAZ"},
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}

	if len(rv) != 1 {
		t.Fail()
	}

	if rv[0].Info == nil {
		t.Fail()
	}

	if len(rv[0].Info.Env) != 1 {
		t.Fail()
	}
}

func TestDockerClientMode(t *testing.T) {
	t.Parallel()

	cm := &clientMock{}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.InContainer),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	if cli.Mode() != "in-container" {
		t.Fail()
	}

	cli, err = client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.LinuxNsenter),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	if cli.Mode() != "linux-nsenter" {
		t.Fail()
	}

	_, err = client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
	)
	if err == nil {
		t.Fail()
	}
}

func TestDockerClientNsEnterInspectError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test-err")

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		cm.Err = testErr

		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.LinuxNsenter),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)

	if !errors.Is(err, testErr) {
		t.Fail()
	}
}

func TestDockerClientNsEnterConnectionsError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test-err")

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cm.OnContainerTop = func() (rv container.ContainerTopOKBody) {
		rv.Titles = []string{"PID"}
		rv.Processes = [][]string{
			{"1"},
		}

		return rv
	}

	failEnter := func(_ int, _ graph.NetProto, _ func(
		_ int, _ *graph.Connection,
	)) error {
		return testErr
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.LinuxNsenter),
		client.WithNsenterFn(failEnter),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)

	if !errors.Is(err, testErr) {
		t.Fail()
	}
}

func TestDockerClientNsEnterContainerTopVariants(t *testing.T) {
	t.Parallel()

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cm.OnContainerTop = func() (rv container.ContainerTopOKBody) {
		rv.Titles = []string{"PID"}
		rv.Processes = [][]string{
			{},
			{"a"},
			{"1"},
		}

		return rv
	}

	var count int

	enter := func(_ int, _ graph.NetProto, _ func(
		_ int, _ *graph.Connection,
	)) error {
		count++

		return nil
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.LinuxNsenter),
		client.WithNsenterFn(enter),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	if _, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	); err != nil {
		t.Fatal()
	}

	if count != 1 {
		t.Fail()
	}
}

func TestDockerClientNsEnterOk(t *testing.T) {
	t.Parallel()

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
						"empty-id": {
							IPAddress: "1.1.1.2",
						},
					},
				},
			},
		}
	}

	cm.OnContainerTop = func() (rv container.ContainerTopOKBody) {
		rv.Titles = []string{"PID"}
		rv.Processes = [][]string{
			{"1"},
		}

		return rv
	}

	testEnter := func(_ int, _ graph.NetProto, fn func(
		_ int, _ *graph.Connection,
	)) error {
		fn(1, &graph.Connection{})

		return nil
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.LinuxNsenter),
		client.WithNsenterFn(testEnter),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	_, err = cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}
}

func TestDockerClientNsEnterLocal(t *testing.T) {
	t.Parallel()

	cm := &clientMock{}

	cm.OnList = func() (rv []types.Container) {
		return []types.Container{
			{
				ID:    "1",
				Names: []string{"test"},
				Image: "test-image",
				State: "running",
				NetworkSettings: &types.SummaryNetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"test-net": {
							EndpointID: "1",
							IPAddress:  "1.1.1.1",
						},
					},
				},
			},
		}
	}

	cm.OnContainerTop = func() (rv container.ContainerTopOKBody) {
		rv.Titles = []string{"PID"}
		rv.Processes = [][]string{
			{"1"},
		}

		return rv
	}

	testEnter := func(_ int, _ graph.NetProto, fn func(int, *graph.Connection)) error {
		loc := net.ParseIP("127.0.0.1")
		nod := net.ParseIP("1.1.1.1")
		rem := net.ParseIP("2.2.2.2")

		fn(1, &graph.Connection{Process: "1", SrcPort: 1, DstPort: 0, SrcIP: nod, Proto: graph.TCP})
		fn(1, &graph.Connection{Process: "1", SrcPort: 10, DstPort: 2, SrcIP: nod, DstIP: rem, Proto: graph.TCP})
		fn(1, &graph.Connection{Process: "1", SrcPort: 5, SrcIP: loc, Proto: graph.TCP})

		return nil
	}

	cli, err := client.NewDocker(
		client.WithClientCreator(func() (client.DockerClient, error) {
			return cm, nil
		}),
		client.WithMode(client.LinuxNsenter),
		client.WithNsenterFn(testEnter),
	)
	if err != nil {
		t.Fatal("client:", err)
	}

	rv, err := cli.Containers(
		context.Background(),
		graph.ALL,
		false,
		nil,
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}

	if len(rv) != 1 {
		t.Fail()
	}

	if rv[0].ConnectionsCount() != 2 {
		t.Fail()
	}

	rv, err = cli.Containers(
		context.Background(),
		graph.ALL,
		true,
		nil,
		voidProgress,
	)
	if err != nil {
		t.Fatal("containers:", err)
	}

	if len(rv) != 1 {
		t.Fail()
	}

	if rv[0].ConnectionsCount() != 3 {
		t.Fail()
	}
}
