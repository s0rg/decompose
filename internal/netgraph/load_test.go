package netgraph_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/s0rg/decompose/internal/netgraph"
)

func TestLoaderLoadError(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("", "", false)
	buf := bytes.NewBufferString(`{`)

	if err := ldr.Load(buf); err == nil {
		t.Fail()
	}
}

func TestLoaderBuildError(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("", "", false)
	buf := bytes.NewBufferString(`{
    "name": "test",
    "is_remote": false,
    "image": "test-image",
    "listen": ["1/tcp", "2/udp"],
    "connected": null
    }`)

	myErr := errors.New("test-error")

	bld := &testBuilder{Err: myErr}

	if err := ldr.Load(buf); err != nil {
		t.Fatal("err=", err)
	}

	err := ldr.Build(bld)
	if err == nil {
		t.Fail()
	}

	if !errors.Is(err, myErr) {
		t.Fail()
	}
}

func TestLoaderSingle(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("", "", false)
	bld := &testBuilder{}
	buf := bytes.NewBufferString(`{
    "name": "test",
    "is_remote": false,
    "image": "test-image",
    "listen": ["1/tcp", "2/udp"],
    "connected": null
    }`)

	if err := ldr.Load(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 1 || bld.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderBadPorts(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("", "", false)
	bld := &testBuilder{}
	buf := bytes.NewBufferString(`{
    "name": "test",
    "is_remote": true,
    "listen": ["#/a/b", "@/udp", ""],
    "connected": null
    }`)

	if err := ldr.Load(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 1 || bld.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderEdges(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("", "", false)
	bld := &testBuilder{}
	buf := bytes.NewBufferString(`{
    "name": "test1",
    "listen": ["1/tcp"],
    "connected": {"test2":["2/tcp"]}
    }
    {
    "name": "test2",
    "listen": ["2/tcp"],
    "connected": {"test1":["1/tcp"]}
    }`)

	if err := ldr.Load(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 2 || bld.Edges != 2 {
		t.Fail()
	}
}

func TestLoaderSeveral(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("", "", false)
	bld := &testBuilder{}

	if err := ldr.Load(bytes.NewBufferString(`{
    "name": "test1",
    "listen": ["1/tcp"],
    "connected": {"test2":["2/tcp"]}
    }`)); err != nil {
		t.Fatal("load1 err=", err)
	}

	if err := ldr.Load(bytes.NewBufferString(`{
    "name": "test2",
    "listen": ["2/tcp"],
    "connected": {"test1":["1/tcp"]}
    }`)); err != nil {
		t.Fatal("load2 err=", err)
	}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 2 || bld.Edges != 2 {
		t.Fail()
	}
}

func TestLoaderEdgesProto(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("tcp", "", false)
	bld := &testBuilder{}
	buf := bytes.NewBufferString(`{
    "name": "test1",
    "listen": ["1/udp"],
    "connected": {"test2":["2/tcp"]}
    }
    {
    "name": "test2",
    "listen": ["2/tcp"],
    "connected": {"test1":["1/udp"]}
    }`)

	if err := ldr.Load(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 2 || bld.Edges != 1 {
		t.Fail()
	}
}

func TestLoaderEdgesFollowNone(t *testing.T) {
	t.Parallel()

	ldr := netgraph.NewLoader("", "foo", false)
	bld := &testBuilder{}
	buf := bytes.NewBufferString(`{
    "name": "test1",
    "listen": ["1/udp"],
    "connected": {"test2":["2/tcp"]}
    }
    {
    "name": "test2",
    "listen": ["2/tcp"],
    "connected": {"test1":["1/udp"]}
    }`)

	if err := ldr.Load(buf); err != nil {
		t.Fatal("load err=", err)
	}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 0 || bld.Edges != 0 {
		t.Fail()
	}
}

func TestLoaderEdgesFollowOne(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "listen": ["1/udp"],
    "connected": {"test2":["2/tcp"], "test3":["3/udp"]}
    }
    {
    "name": "test2",
    "listen": ["2/tcp"],
    "connected": {"test1":["1/udp"], "test3":["3/tcp"]}
    }
    {
    "name": "test3",
    "listen": ["3/tcp", "3/udp"],
    "connected": {"test1":["1/udp"]}
    }`)

	ldr := netgraph.NewLoader("udp", "test3", false)

	if err := ldr.Load(buf); err != nil {
		t.Fatal("load err=", err)
	}

	bld := &testBuilder{}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 2 || bld.Edges != 2 {
		t.Fail()
	}
}

func TestLoaderLocal(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString(`{
    "name": "test1",
    "listen": ["1/udp"],
    "connected": {"test2":["2/tcp"]}
    }
    {
    "name": "test2",
    "listen": ["2/tcp"],
    "is_external": true,
    "connected": {"test1":["1/udp"]}
    }
    `)

	ldr := netgraph.NewLoader("", "", true)

	if err := ldr.Load(buf); err != nil {
		t.Fatal("load err=", err)
	}

	bld := &testBuilder{}

	if err := ldr.Build(bld); err != nil {
		t.Fatal("build err=", err)
	}

	if bld.Nodes != 1 || bld.Edges != 0 {
		t.Fail()
	}
}
