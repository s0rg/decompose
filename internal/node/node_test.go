package node_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/node"
)

func makeTestNode(
	id, name, image string,
	ports []*node.Port,
) (rv *node.Node) {
	rv = &node.Node{
		ID:    id,
		Name:  name,
		Ports: &node.Ports{},
	}

	for _, p := range ports {
		rv.Ports.Add("", p)
	}

	return rv
}

func TestNodeIsExternal(t *testing.T) {
	t.Parallel()

	n := node.Node{}

	if !n.IsExternal() {
		t.Fail()
	}

	n.ID = "id"

	if n.IsExternal() {
		t.Fail()
	}
}

func TestNodeToJSON(t *testing.T) {
	t.Parallel()

	nodeMeta := makeTestNode("test-id", "test-name", "test-image", []*node.Port{
		{Kind: "udp", Value: 53},
	})

	nodeMeta.Meta = &node.Meta{
		Info: "test",
		Tags: []string{"test"},
	}

	nodeContainer := makeTestNode("test-id", "test-name", "test-image", []*node.Port{
		{Kind: "udp", Value: 53},
	})

	nodeContainer.Container.Cmd = []string{"foo"}
	nodeContainer.Container.Env = []string{"A=B"}

	testCases := []struct {
		Node       *node.Node
		Name       string
		Image      string
		PortsNum   int
		Volumes    int
		External   bool
		HasMeta    bool
		HasProcess bool
	}{
		{
			Node: makeTestNode("test-id", "test-name", "", []*node.Port{
				{Kind: "tcp", Value: 80},
				{Kind: "udp", Value: 53},
			}),
			Name:     "test-name",
			PortsNum: 2,
		},
		{
			Node: makeTestNode("test-id", "test-id", "", []*node.Port{
				{Kind: "tcp", Value: 80},
			}),
			Name:     "test-id",
			PortsNum: 1,
			External: true,
		},
		{
			Node: makeTestNode("test-id", "test-name", "test-image", []*node.Port{
				{Kind: "udp", Value: 53},
			}),
			Name:     "test-name",
			Image:    "test-image",
			PortsNum: 1,
		},
		{
			Node:     nodeMeta,
			Name:     "test-name",
			Image:    "test-image",
			PortsNum: 1,
			HasMeta:  true,
		},
		{
			Node:       nodeContainer,
			Name:       "test-name",
			Image:      "test-image",
			PortsNum:   1,
			HasProcess: true,
		},
		{
			Node: &node.Node{
				ID:    "test-id",
				Name:  "test-name",
				Image: "test-image",
				Ports: &node.Ports{},
				Volumes: []*node.Volume{
					{Type: "none"},
					{Type: "bind"},
				},
			},
			Name:    "test-name",
			Image:   "test-image",
			Volumes: 2,
		},
	}

	for _, tc := range testCases {
		j := tc.Node.ToJSON()

		if j.Name != tc.Name {
			t.Fatal("name", tc)
		}

		if j.IsExternal != tc.External {
			t.Fatal("external", tc)
		}

		if len(j.Listen) != tc.PortsNum {
			t.Fatal("listen", tc)
		}

		if tc.HasMeta && len(j.Tags) == 0 {
			t.Fatal("extra", tc)
		}

		if tc.HasProcess && len(j.Container.Cmd) == 0 {
			t.Fatal("process", tc)
		}

		if len(j.Volumes) != tc.Volumes {
			t.Fatal("volumes", tc)
		}

		if tc.Image == "" {
			continue
		}

		if j.Image == nil {
			t.Fatal("image == nil", tc)
		}

		if *j.Image != tc.Image {
			t.Fatal("image", tc)
		}
	}
}

func TestNodeToView(t *testing.T) {
	t.Parallel()

	nodeContainer := makeTestNode("test-id", "test-name", "test-image", []*node.Port{
		{Kind: "udp", Value: 53},
	})

	nodeContainer.Container.Cmd = []string{"foo", "-arg"}

	testCases := []struct {
		Node     *node.Node
		Cmd      string
		Tags     int
		Args     int
		External bool
	}{
		{
			Node: &node.Node{
				Name:  "test",
				Image: "image",
			},
			External: false,
			Tags:     0,
			Cmd:      "",
			Args:     0,
		},
		{
			Node: &node.Node{
				Name:  "test",
				Image: "image",
				Meta: &node.Meta{
					Info: "",
				},
			},
			External: false,
			Tags:     0,
			Cmd:      "",
			Args:     0,
		},
		{
			Node: &node.Node{
				Name:  "test",
				Image: "image",
				Meta: &node.Meta{
					Info: "",
					Tags: []string{"a"},
				},
			},
			External: false,
			Tags:     1,
			Cmd:      "",
			Args:     0,
		},
		{
			Node: &node.Node{
				ID:   "test",
				Name: "test",
			},
			External: true,
			Tags:     0,
			Cmd:      "",
			Args:     0,
		},
		{
			Node:     nodeContainer,
			External: true,
			Tags:     0,
			Cmd:      "foo",
			Args:     1,
		},
	}

	for _, tc := range testCases {
		v := tc.Node.ToView()

		if v.Name != tc.Node.Name {
			t.Fail()
		}

		if v.Image != tc.Node.Image {
			t.Fail()
		}

		if v.IsExternal != tc.External {
			t.Fail()
		}

		if v.Cmd != tc.Cmd {
			t.Fail()
		}

		if len(v.Tags) != tc.Tags {
			t.Fail()
		}

		if len(v.Args) != tc.Args {
			t.Fail()
		}
	}
}
