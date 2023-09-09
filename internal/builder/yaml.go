//go:build !test

// i/o here, nothing to test

package builder

import (
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/s0rg/decompose/internal/node"
	"github.com/s0rg/set"
)

const volumeSuffix = "_data"

type compose struct {
	Services map[string]*service `yaml:"services"`
	Networks map[string]*network `yaml:"networks"`
	Volumes  map[string]any      `yaml:"volumes"`
}

type network struct {
	External bool `yaml:"external"`
}

type service struct {
	Image       string    `yaml:"image"`
	Expose      yaml.Node `yaml:"expose"`
	Links       []string  `yaml:"links"`
	Volumes     []string  `yaml:"volumes"`
	Networks    []string  `yaml:"networks"`
	Environment yaml.Node `yaml:"environment"`
	Command     []string  `yaml:"command"`
}

type YAML struct {
	state *compose
	idmap map[string]string
}

func NewYAML() *YAML {
	return &YAML{
		state: &compose{
			Services: make(map[string]*service),
			Networks: make(map[string]*network),
			Volumes:  make(map[string]any),
		},
		idmap: make(map[string]string),
	}
}

func (y *YAML) Name() string {
	return "compose-yaml"
}

func (y *YAML) AddNode(n *node.Node) error {
	if n.IsExternal() {
		return nil
	}

	svc := &service{
		Image:    n.Image,
		Networks: n.Networks,
	}

	if n.Process != nil {
		if len(n.Process.Cmd) > 0 {
			svc.Command = n.Process.Cmd
		}

		if len(n.Process.Env) > 0 {
			yn := yaml.Node{
				Kind: yaml.SequenceNode,
			}

			for _, ev := range n.Process.Env {
				yn.Content = append(yn.Content, &yaml.Node{
					Kind:  yaml.ScalarNode,
					Style: yaml.DoubleQuotedStyle,
					Value: ev,
				})
			}

			svc.Environment = yn
		}
	}

	for _, name := range n.Networks {
		y.state.Networks[name] = &network{
			External: true,
		}
	}

	svc.Expose = yaml.Node{
		Kind: yaml.SequenceNode,
	}

	for _, p := range n.Ports {
		svc.Expose.Content = append(svc.Expose.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Style: yaml.DoubleQuotedStyle,
			Value: p.Label(),
		})
	}

	svc.Volumes = make([]string, len(n.Volumes))

	for i, v := range n.Volumes {
		src := v.Src

		if v.Type == "volume" {
			src = strings.ReplaceAll(n.Name, "-", "_") + volumeSuffix
			y.state.Volumes[src] = nil
		}

		svc.Volumes[i] = src + ":" + v.Dst
	}

	y.idmap[n.ID] = n.Name
	y.state.Services[n.Name] = svc

	return nil
}

func (y *YAML) AddEdge(srcID, dstID string, _ *node.Port) {
	name, ok := y.idmap[srcID]
	if !ok {
		return
	}

	svc, ok := y.state.Services[name]
	if !ok {
		return
	}

	name, ok = y.idmap[dstID]
	if !ok {
		return
	}

	if _, ok = y.state.Services[name]; !ok {
		return
	}

	svc.Links = append(svc.Links, name)
}

func (y *YAML) Write(w io.Writer) {
	enc := yaml.NewEncoder(w)
	defer enc.Close()

	s := make(set.Unordered[string])

	// de-dup links
	for _, svc := range y.state.Services {
		set.Load(s, svc.Links...)
		svc.Links = set.ToSlice(s)
		s.Clear()
	}

	_ = enc.Encode(y.state)
}
