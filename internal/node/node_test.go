package node_test

import (
	"testing"

	"github.com/s0rg/decompose/internal/node"
)

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

	testCases := []struct {
		Node     *node.Node
		Name     string
		Image    string
		PortsNum int
		External bool
	}{
		{
			Node: &node.Node{
				ID:   "test-id",
				Name: "test-name",
				Ports: []node.Port{
					{Kind: "tcp", Value: 80},
					{Kind: "udp", Value: 53},
				},
			},
			Name:     "test-name",
			PortsNum: 2,
			External: false,
		},
		{
			Node: &node.Node{
				ID:   "test-id",
				Name: "test-id",
				Ports: []node.Port{
					{Kind: "tcp", Value: 80},
				},
			},
			Name:     "test-id",
			PortsNum: 1,
			External: true,
		},
		{
			Node: &node.Node{
				ID:    "test-id",
				Name:  "test-name",
				Image: "test-image",
				Ports: []node.Port{
					{Kind: "udp", Value: 53},
				},
			},
			Name:     "test-name",
			Image:    "test-image",
			PortsNum: 1,
			External: false,
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
