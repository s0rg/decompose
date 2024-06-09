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
		Image: image,
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
		{Kind: "udp", Value: "53"},
	})

	nodeMeta.Meta = &node.Meta{
		Info: "test",
		Tags: []string{"test"},
	}

	nodeContainer := makeTestNode("test-id2", "test-name", "test-image", []*node.Port{
		{Kind: "udp", Value: "53"},
	})

	nodeContainer.Container.Cmd = []string{"foo"}
	nodeContainer.Container.Env = []string{"A=B"}

	testCases := []struct {
		Node       *node.Node
		Name       string
		Image      string
		Volumes    int
		External   bool
		HasMeta    bool
		HasProcess bool
	}{
		{
			Node: makeTestNode("test-id", "test-name1", "", []*node.Port{
				{Kind: "tcp", Value: "80"},
				{Kind: "udp", Value: "53"},
			}),
			Name: "test-name1",
		},
		{
			Node: makeTestNode("test-id", "test-id", "", []*node.Port{
				{Kind: "tcp", Value: "80"},
			}),
			Name:     "test-id",
			External: true,
		},
		{
			Node: makeTestNode("test-id3", "test-name", "test-image", []*node.Port{
				{Kind: "udp", Value: "53"},
			}),
			Name:  "test-name",
			Image: "test-image",
		},
		{
			Node:    nodeMeta,
			Name:    "test-name",
			Image:   "test-image",
			HasMeta: true,
		},
		{
			Node:       nodeContainer,
			Name:       "test-name",
			Image:      "test-image",
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
		{Kind: "udp", Value: "53"},
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
			External: false,
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

func TestNodeFormatMeta(t *testing.T) {
	t.Parallel()

	n := node.Node{}

	if _, ok := n.FormatMeta(); ok {
		t.Fail()
	}

	n.Meta = &node.Meta{
		Info: "foo",
		Docs: "bar",
		Repo: "baz",
		Tags: []string{"a", "b"},
	}

	if _, ok := n.FormatMeta(); !ok {
		t.Fail()
	}
}
