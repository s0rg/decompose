package graph_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/graph"
)

var testCases = []struct {
	Conns     []*graph.Connection
	Listeners int
	Outbounds int
	Count     int
}{
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
			{LocalPort: 1, RemotePort: 0}, // listener
			{LocalPort: 2, RemotePort: 1}, // outbound
		},
		Listeners: 1,
		Outbounds: 1,
		Count:     2,
	},
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
			{LocalPort: 2, RemotePort: 1}, // outbound
		},
		Listeners: 0,
		Outbounds: 1,
		Count:     1,
	},
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
			{LocalPort: 1, RemotePort: 0}, // listener
		},
		Listeners: 1,
		Outbounds: 0,
		Count:     1,
	},
	{
		Conns: []*graph.Connection{
			{LocalPort: 1, RemotePort: 2}, // inbound
		},
		Listeners: 0,
		Outbounds: 0,
		Count:     0,
	},
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

func TestContainerCount(t *testing.T) {
	t.Parallel()

	for i := 0; i < len(testCases); i++ {
		tc := &testCases[i]

		c := graph.Container{}
		c.SetConnections(tc.Conns)

		if res := c.ConnectionsCount(); res != tc.Count {
			t.Fatalf("test case[%d] fail want %d got %d", i, tc.Count, res)
		}
	}
}
