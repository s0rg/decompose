package graph_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/s0rg/set"

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
	_ []string,
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

func (tb *testBuilder) AddNode(n *node.Node) error {
	log.Printf("%+v", n)

	if tb.Err != nil {
		return tb.Err
	}

	tb.Nodes++

	return nil
}

func (tb *testBuilder) AddEdge(e *node.Edge) {

	log.Printf("%+v", e)
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
		Info: &graph.ContainerInfo{
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
	cli.Data[0].AddMany([]*graph.Connection{
		{SrcPort: 1, Proto: graph.TCP, Listen: true},                 // listen 1
		{DstIP: node2, SrcPort: 10, DstPort: 2, Proto: graph.TCP},    // connected to node2:2
		{DstIP: external, SrcPort: 10, DstPort: 1, Proto: graph.TCP}, // connected to external:1
	})

	// node 2
	cli.Data[1].AddMany([]*graph.Connection{
		{SrcPort: 2, Proto: graph.TCP, Listen: true},                 // listen 2
		{DstIP: node3, SrcPort: 10, DstPort: 3, Proto: graph.TCP},    // connected to node3:3
		{DstIP: external, SrcPort: 10, DstPort: 2, Proto: graph.TCP}, // connected to external:2
	})

	// node 3
	cli.Data[2].AddMany([]*graph.Connection{
		{SrcPort: 3, Proto: graph.TCP, Listen: true},                 // listen 3
		{DstIP: node1, SrcPort: 10, DstPort: 1, Proto: graph.TCP},    // connected to node1:1
		{DstIP: node2, SrcPort: 122, DstPort: 22, Proto: graph.TCP},  // connected to node2:22
		{DstIP: external, SrcPort: 10, DstPort: 3, Proto: graph.TCP}, // connected to external:3
	})

	return cli
}

func TestBuildSimple(t *testing.T) {
	t.Parallel()

	cli := testClientWithEnv()
	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder: bld,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 4 || bld.Edges != 7 {
		t.Fail()
	}
}

func TestBuildFollow(t *testing.T) {
	t.Parallel()

	flw := make(set.Unordered[string])
	flw.Add("1")

	cli := testClientWithEnv()
	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder: bld,
		Follow:  flw,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 3 || bld.Edges != 2 {
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
		Meta:      ext,
		Proto:     graph.ALL,
		OnlyLocal: true,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 3 || bld.Edges != 4 {
		t.Fail()
	}
}

func TestBuildNoNodes(t *testing.T) {
	t.Parallel()

	flw := make(set.Unordered[string])
	flw.Add("5")

	cli := testClientWithEnv()
	cfg := &graph.Config{
		Follow:    flw,
		OnlyLocal: true,
		Proto:     graph.ALL,
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
		Builder: bld,
		Meta:    ext,
		Proto:   graph.ALL,
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
		Builder: bld,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildLoops(t *testing.T) {
	t.Parallel()

	node1 := net.ParseIP("1.1.1.1")
	node2 := net.ParseIP("1.1.1.2")
	local := net.ParseIP("127.0.0.1")

	cli := &testClient{Data: []*graph.Container{
		makeContainer("1", node1.String()),
		makeContainer("2", node2.String()),
	}}

	cli.Data[0].AddMany([]*graph.Connection{
		{SrcPort: 1, Proto: graph.TCP},                            // listen 1
		{DstIP: node2, SrcPort: 10, DstPort: 2, Proto: graph.TCP}, // connected to node2:2
		{DstIP: node1, SrcPort: 10, DstPort: 1, Proto: graph.TCP}, // connected to itself
	})

	cli.Data[1].AddMany([]*graph.Connection{
		{SrcPort: 2, Proto: graph.TCP},                            // listen 2
		{DstIP: node1, SrcPort: 10, DstPort: 1, Proto: graph.TCP}, // connected to node1:1
		{DstIP: local, SrcPort: 11, DstPort: 2, Proto: graph.TCP}, // connected to self:2
	})

	bld := &testBuilder{}
	ext := &testEnricher{}
	cfg := &graph.Config{
		Builder:   bld,
		Meta:      ext,
		Proto:     graph.ALL,
		OnlyLocal: true,
	}

	if err := graph.Build(cfg, cli); err != nil {
		t.Fatalf("err = %v", err)
	}

	if bld.Nodes != 2 || bld.Edges != 4 {
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
