package graph_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
	"github.com/s0rg/decompose/internal/node"
)

type testClient struct {
	Err  error
	Data []*graph.Container
}

func (tc *testClient) Containers(
	_ context.Context,
	_ graph.NetProto,
	_ bool,
	fn func(int, int),
) ([]*graph.Container, error) {
	if tc.Err != nil {
		return nil, tc.Err
	}

	l := len(tc.Data)

	fn(0, l)

	if l > 1 {
		fn(l/2, l)
	}

	fn(l, l)

	return tc.Data, nil
}

type testBuilder struct {
	Err   error
	Nodes int
	Edges int
}

func (tb *testBuilder) AddNode(_ *node.Node) error {
	if tb.Err != nil {
		return tb.Err
	}

	tb.Nodes++

	return nil
}

func (tb *testBuilder) AddEdge(_, _ string, _ node.Port) {
	tb.Edges++
}

func (tb *testBuilder) Reset() {
	tb.Nodes, tb.Edges = 0, 0
}

type testEnricher struct{}

func (de *testEnricher) Enrich(_ *node.Node) {}

func TestBuildError(t *testing.T) {
	t.Parallel()

	myErr := errors.New("test error")
	cli := &testClient{Err: myErr}
	cfg := &graph.Config{
		Proto: graph.ALL,
	}

	err := graph.Build(cfg, cli)
	if err == nil {
		t.Fatal("err is nil")
	}

	if !errors.Is(err, myErr) {
		t.Fatalf("unknown error, want: %v got: %v", myErr, err)
	}
}

func TestBuildOneConainer(t *testing.T) {
	t.Parallel()

	cli := &testClient{Data: []*graph.Container{
		{},
	}}

	cfg := &graph.Config{
		Proto: graph.ALL,
	}

	if err := graph.Build(cfg, cli); err == nil {
		t.Fail()
	}
}

func makeContainer(name, ip string) *graph.Container {
	return &graph.Container{
		ID:    name + "-id",
		Name:  name,
		Image: name + "-image:latest",
		Endpoints: map[string]string{
			ip: "test-net",
		},
		Process: &graph.ProcessInfo{
			Cmd: []string{"test-app", "-test-arg"},
			Env: []string{"FOO=BAR"},
		},
		Volumes: []*graph.VolumeInfo{
			{Type: "bind"},
		},
	}
}

func testClientWithEnv() graph.ContainerClient {
	node1 := net.ParseIP("1.1.1.1")
	node2 := net.ParseIP("1.1.1.2")
	node3 := net.ParseIP("1.1.1.3")
	external := net.ParseIP("2.2.2.1")

	cli := &testClient{Data: []*graph.Container{
		makeContainer("1", node1.String()),
		makeContainer("2", node2.String()),
		makeContainer("3", node3.String()),
	}}

	// node 1
	cli.Data[0].SetConnections([]*graph.Connection{
		{LocalPort: 1, Proto: graph.TCP},                                     // listen 1
		{RemoteIP: node2, LocalPort: 10, RemotePort: 2, Proto: graph.TCP},    // connected to node2:2
		{RemoteIP: external, LocalPort: 10, RemotePort: 1, Proto: graph.TCP}, // connected to external:1
	})

	// node 2
	cli.Data[1].SetConnections([]*graph.Connection{
		{LocalPort: 2, Proto: graph.TCP},                                     // listen 2
		{RemoteIP: node3, LocalPort: 10, RemotePort: 3, Proto: graph.TCP},    // connected to node3:3
		{RemoteIP: external, LocalPort: 10, RemotePort: 2, Proto: graph.TCP}, // connected to external:2
	})

	// node 3
	cli.Data[2].SetConnections([]*graph.Connection{
		{LocalPort: 3, Proto: graph.TCP},                                     // listen 3
		{RemoteIP: node1, LocalPort: 10, RemotePort: 1, Proto: graph.TCP},    // connected to node1:1
		{RemoteIP: external, LocalPort: 10, RemotePort: 3, Proto: graph.TCP}, // connected to external:3
	})

	return cli
}

func TestBuildSimple(t *testing.T) {
	t.Parallel()

	cli := testClientWithEnv()
	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder:  bld,
		Enricher: ext,
		Proto:    graph.ALL,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 4 || bld.Edges != 6 {
		t.Fail()
	}
}

func TestBuildFollow(t *testing.T) {
	t.Parallel()

	cli := testClientWithEnv()
	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder:  bld,
		Enricher: ext,
		Proto:    graph.ALL,
		Follow:   "1",
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 3 || bld.Edges != 3 {
		t.Fail()
	}
}

func TestBuildLocal(t *testing.T) {
	t.Parallel()

	cli := testClientWithEnv()
	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder:   bld,
		Enricher:  ext,
		Proto:     graph.ALL,
		OnlyLocal: true,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 3 || bld.Edges != 3 {
		t.Fail()
	}
}

func TestBuildNoNodes(t *testing.T) {
	t.Parallel()

	cli := testClientWithEnv()
	cfg := &graph.Config{
		Proto:  graph.ALL,
		Follow: "4",
	}

	if err := graph.Build(cfg, cli); err == nil {
		t.Fail()
	}
}

func TestBuildNodeError(t *testing.T) {
	t.Parallel()

	myErr := errors.New("test error")
	cli := testClientWithEnv()
	bld := &testBuilder{Err: myErr}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder:  bld,
		Enricher: ext,
		Proto:    graph.ALL,
	}

	err := graph.Build(cfg, cli)
	if err == nil {
		t.Fatal("err is nil")
	}

	if !errors.Is(err, myErr) {
		t.Fatalf("unknown error, want: %v got: %v", myErr, err)
	}
}

func TestBuildNoConnections(t *testing.T) {
	t.Parallel()

	cli := &testClient{Data: []*graph.Container{
		makeContainer("1", "1.1.1.1"),
		makeContainer("2", "2.2.2.2"),
		makeContainer("3", "3.3.3.3"),
	}}

	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder:  bld,
		Enricher: ext,
		Proto:    graph.ALL,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildLoops(t *testing.T) {
	t.Parallel()

	node1 := net.ParseIP("1.1.1.1")
	node2 := net.ParseIP("1.1.1.2")

	cli := &testClient{Data: []*graph.Container{
		makeContainer("1", node1.String()),
		makeContainer("2", node2.String()),
	}}

	cli.Data[0].SetConnections([]*graph.Connection{
		{LocalPort: 1, Proto: graph.TCP},                                  // listen 1
		{RemoteIP: node2, LocalPort: 10, RemotePort: 2, Proto: graph.TCP}, // connected to node2:2
		{RemoteIP: node1, LocalPort: 10, RemotePort: 1, Proto: graph.TCP}, // connected to itself
	})

	cli.Data[1].SetConnections([]*graph.Connection{
		{LocalPort: 2, Proto: graph.TCP},                                  // listen 2
		{RemoteIP: node1, LocalPort: 10, RemotePort: 1, Proto: graph.TCP}, // connected to node1:1
	})

	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder:   bld,
		Enricher:  ext,
		Proto:     graph.ALL,
		OnlyLocal: true,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 2 || bld.Edges != 3 {
		t.Fail()
	}

	bld.Reset()

	cfg.NoLoops = true

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 2 || bld.Edges != 2 {
		t.Fail()
	}
}
