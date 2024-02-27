package graph_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/s0rg/set"

	"github.com/s0rg/decompose/internal/graph"
)

func TestLoaderLoadError(t *testing.T) {
	t.Parallel()

	ext := &testEnricher{}
	cfg := &graph.Config{
		Proto: graph.ALL,
		Meta:  ext,
	}

	ldr := graph.NewLoader(cfg)
	buf := bytes.NewBufferString(`{`)

	if err := ldr.FromReader(buf); err == nil {
		t.Fail()
	}
}

func TestLoaderBuildError(t *testing.T) {
	t.Parallel()

	myErr := errors.New("test-error")
	bldr := &testBuilder{Err: myErr}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	buf := bytes.NewBufferString(`{
    "name": "test",
    "is_remote": false,
    "image": "test-image",
    "listen": {"foo":["1/tcp", "2/udp"]},
    "connected": null
    }`)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("err=", err)
	}

	err := ldr.Build()
	if err == nil {
		t.Fail()
	}

	if !errors.Is(err, myErr) {
		t.Fail()
	}
}

func TestLoaderSingle(t *testing.T) {
	t.Parallel()

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	buf := bytes.NewBufferString(`{
    "name": "test",
    "is_remote": false,
    "image": "test-image",
    "listen": {"foo": ["1/tcp", "2/udp"]},
    "connected": null
    }`)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 1 || bldr.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderBadPorts(t *testing.T) {
	t.Parallel()

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	buf := bytes.NewBufferString(`{
    "name": "test",
    "is_remote": true,
    "listen": {"foo": ["#/a/b", "@/udp", ""]},
    "connected": null
    }`)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 1 || bldr.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderEdges(t *testing.T) {
	t.Parallel()

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "listen": {"foo": ["1/tcp"]},
    "connected": {"test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}]}
    }
    {
    "name": "test2",
    "listen": {"bar":["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/tcp"}]}
    }`)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 2 || bldr.Edges != 2 {
		t.Fail()
	}
}

func TestLoaderSeveral(t *testing.T) {
	t.Parallel()

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(bytes.NewBufferString(`{
    "name": "test1",
    "networks": ["foo"],
    "listen": {"foo": ["1/tcp"]},
    "connected": {"test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}]}
    }`)); err != nil {
		t.Fatal("load1 err=", err)
	}

	if err := ldr.FromReader(bytes.NewBufferString(`{
    "name": "test2",
    "networks": ["foo"],
    "listen": {"bar":["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/tcp"}]}
    }`)); err != nil {
		t.Fatal("load2 err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 2 || bldr.Edges != 2 {
		t.Fail()
	}
}

func TestLoaderEdgesProto(t *testing.T) {
	t.Parallel()

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.TCP,
	}

	ldr := graph.NewLoader(cfg)

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "networks": ["foo"],
    "listen": {"foo": ["1/udp"]},
    "connected": {"test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}]}
    }
    {
    "name": "test2",
    "networks": ["foo"],
    "listen": {"bar":["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/udp"}]}
    }`)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 2 || bldr.Edges != 1 {
		t.Fail()
	}
}

func TestLoaderEdgesFollowNone(t *testing.T) {
	t.Parallel()

	bldr := &testBuilder{}
	ext := &testEnricher{}

	flw := make(set.Unordered[string])
	flw.Add("foo")

	cfg := &graph.Config{
		Builder: bldr,
		Follow:  flw,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "networks": ["foo"],
    "listen": {"foo": ["1/udp"]},
    "connected": {"test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}]}
    }
    {
    "name": "test2",
    "networks": ["foo"],
    "listen": {"bar":["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/udp"}]}
    }`)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 0 || bldr.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderEdgesFollowOne(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "networks": ["foo"],
    "listen": {"foo": ["1/udp"]},
    "connected": {
        "test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}],
        "test3":[{"src": "foo", "dst": "baz", "port": "3/udp"}]
      }
    }
    {
    "name": "test2",
    "networks": ["foo"],
    "listen": {"bar":["2/tcp"]},
    "connected": {
        "test1":[{"src": "bar", "dst": "foo", "port": "1/udp"}]
      }
    }
    {
    "name": "test3",
    "networks": ["foo"],
    "listen": {"bar":["3/tcp", "3/udp"]},
    "connected": {
        "test1":[{"src": "baz", "dst": "foo", "port": "1/udp"}]
      }
    }`)

	bldr := &testBuilder{}
	ext := &testEnricher{}

	flw := make(set.Unordered[string])
	flw.Add("test3")

	cfg := &graph.Config{
		Builder: bldr,
		Follow:  flw,
		Meta:    ext,
		Proto:   graph.UDP,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 2 || bldr.Edges != 2 {
		t.Fail()
	}
}

func TestLoaderLocal(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "networks": ["foo"],
    "listen": {"foo": ["1/udp"]},
    "connected": {"test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}]}
    }
    {
    "name": "test2",
    "is_external": true,
    "networks": ["foo"],
    "listen": {"bar": ["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/udp"}]}
    }
    `)

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder:   bldr,
		Meta:      ext,
		OnlyLocal: true,
		Proto:     graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 1 || bldr.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderMeta(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "tags": ["test"],
    "networks": ["foo"],
    "listen": {"foo": ["1/udp"]},
    "connected": {"test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}]}
    }
    {
    "name": "test2",
    "is_external": true,
    "networks": ["foo"],
    "listen": {"bar": ["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/udp"}]}
    }`)

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder:   bldr,
		Meta:      ext,
		OnlyLocal: true,
		Proto:     graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 1 || bldr.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderFull(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "container": {"cmd": ["foo", "bar"], "env": ["A=B"], "labels": {}},
    "volumes": [{"type": "bind", "src": "", "dst": ""}],
    "networks": ["foo"],
    "listen": {"foo": ["1/udp"]},
    "connected": {"test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}]}
    }
    {
    "name": "test2",
    "is_external": true,
    "networks": ["foo"],
    "listen": {"bar": ["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/udp"}]}
    }`)

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder:  bldr,
		Meta:     ext,
		FullInfo: true,
		Proto:    graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 2 || bldr.Edges != 2 {
		t.Fail()
	}
}

func TestLoaderLoops(t *testing.T) {
	t.Parallel()

	const rawJSON = `{
    "name": "test1",
    "networks": ["foo"],
    "listen": {"foo": ["1/udp"]},
    "connected": {
        "test2":[{"src": "foo", "dst": "bar", "port": "2/tcp"}],
        "test1":[{"src": "foo", "dst": "foo", "port": "1/udp"}]
      }
    }
    {
    "name": "test2",
    "is_external": true,
    "networks": ["foo"],
    "listen": {"bar": ["2/tcp"]},
    "connected": {"test1":[{"src": "bar", "dst": "foo", "port": "1/udp"}]}
    }`

	bldr := &testBuilder{}
	ext := &testEnricher{}

	cfg := &graph.Config{
		Builder: bldr,
		Meta:    ext,
		Proto:   graph.ALL,
	}

	ldr := graph.NewLoader(cfg)

	if err := ldr.FromReader(bytes.NewBufferString(rawJSON)); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 2 || bldr.Edges != 3 {
		t.Fail()
	}

	bldr.Reset()

	cfg.NoLoops = true
	ldr = graph.NewLoader(cfg)

	if err := ldr.FromReader(bytes.NewBufferString(rawJSON)); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(); err != nil {
		t.Fatal("build err=", err)
	}

	if bldr.Nodes != 2 || bldr.Edges != 2 {
		t.Fail()
	}
}
