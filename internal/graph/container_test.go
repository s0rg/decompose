package graph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/graph"
)

var testCases = []struct {
	Conns     []*graph.Connection
	Listeners int
	Outbounds int
}{
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
			{LocalPort: 1, RemotePort: 0}, // listener
			{LocalPort: 2, RemotePort: 1}, // outbound
		},
		Listeners: 1,
		Outbounds: 1,
	},
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
			{LocalPort: 2, RemotePort: 1}, // outbound
		},
		Listeners: 0,
		Outbounds: 1,
	},
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
			{LocalPort: 1, RemotePort: 0}, // listener
		},
		Listeners: 1,
		Outbounds: 0,
	},
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
		},
		Listeners: 0,
		Outbounds: 0,
	},
}

func TestContainerMatch(t *testing.T) {
	t.Parallel()

	c := graph.Container{}

	const id = "test-id"

	if !c.Match("") { // empty string match all
		t.Fail()
	}

	c.ID = id

	if c.Match(id) { // by id
		t.Fail()
	}

	c.Name = id

	if !c.Match(id) { // by name
		t.Fail()
	}
}

func TestContainerListeners(t *testing.T) {
	t.Parallel()

	for i := 0; i < len(testCases); i++ {
		tc := &testCases[i]

		res := 0

		c := graph.Container{}
		c.SetConnections(tc.Conns)
		c.ForEachListener(func(_ *graph.Connection) {
			res++
		})

		if res != tc.Listeners {
			t.Fatalf("test case[%d] fail want %d got %d", i, tc.Listeners, res)
		}
	}
}

func TestContainerOutbounds(t *testing.T) {
	t.Parallel()

	for i := 0; i < len(testCases); i++ {
		tc := &testCases[i]

		res := 0

		c := graph.Container{}
		c.SetConnections(tc.Conns)
		c.ForEachOutbound(func(_ *graph.Connection) {
			res++
		})

		if res != tc.Outbounds {
			t.Fatalf("test case[%d] fail want %d got %d", i, tc.Outbounds, res)
		}
	}
}
