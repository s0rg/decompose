package graph

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidRange  = errors.New("invalid range")
	ErrPortCollision = errors.New("ports collision")
)

type (
	ruleJSON struct {
		Name   string   `json:"name"`
		Ports  []string `json:"ports"`
		Weight int      `json:"weight"`
	}

	ClusterBuilder struct {
		builder NamedBuilderWriter
		weights map[string]int
		nodes   map[string]*node.Node
		index   map[string]map[int]string
		cluster map[string]map[string]node.Ports
	}
)

func NewClusterBuilder(b NamedBuilderWriter) *ClusterBuilder {
	return &ClusterBuilder{
		builder: b,
		weights: make(map[string]int),
		nodes:   make(map[string]*node.Node),
		index:   make(map[string]map[int]string),
		cluster: make(map[string]map[string]node.Ports),
	}
}

func (cb *ClusterBuilder) Name() string {
	return cb.builder.Name() + " clustered"
}

func (cb *ClusterBuilder) Write(w io.Writer) {
	for src, dmap := range cb.cluster {
		for dst, ports := range dmap {
			for _, p := range ports.Dedup() {
				cb.builder.AddEdge(src, dst, p)
			}
		}
	}

	cb.builder.Write(w)
}

func (cb *ClusterBuilder) AddNode(n *node.Node) error {
	if cluster, ok := cb.Match(n); ok {
		n.Cluster = cluster
	}

	cb.nodes[n.ID] = n

	if err := cb.builder.AddNode(n); err != nil {
		return fmt.Errorf("builder: %w", err)
	}

	return nil
}

func (cb *ClusterBuilder) AddEdge(src, dst string, port node.Port) {
	nsrc, ok := cb.nodes[src]
	if !ok {
		return
	}

	ndst, ok := cb.nodes[dst]
	if !ok {
		return
	}

	if nsrc.Cluster != "" && ndst.Cluster != "" && nsrc.Cluster != ndst.Cluster {
		if cluster, ok := cb.clusterFor(&port); ok && ndst.Cluster == cluster {
			cdst, ok := cb.cluster[nsrc.Cluster]
			if !ok {
				cdst = make(map[string]node.Ports)
			}

			cdst[ndst.Cluster] = append(cdst[ndst.Cluster], port)
			cb.cluster[nsrc.Cluster] = cdst

			return
		}
	}

	cb.builder.AddEdge(src, dst, port)
}

func (cb *ClusterBuilder) FromReader(r io.Reader) (err error) {
	var rules []ruleJSON

	dec := json.NewDecoder(r)

	for dec.More() {
		if err = dec.Decode(&rules); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}

	for i := 0; i < len(rules); i++ {
		rule := &rules[i]

		weight := rule.Weight
		if weight == 0 {
			weight = 1
		}

		cb.weights[rule.Name] = weight

		for j := 0; j < len(rule.Ports); j++ {
			proto, ports, perr := parseRulePorts(rule.Ports[j])
			if perr != nil {
				return fmt.Errorf("parse '%s': %w", rule.Ports[j], perr)
			}

			pmap, ok := cb.index[proto]
			if !ok {
				pmap = make(map[int]string)
			}

			for k := 0; k < len(ports); k++ {
				port := ports[k]

				if exist, ok := pmap[port]; ok {
					return fmt.Errorf("%w %s - %s", ErrPortCollision, rule.Name, exist)
				}

				pmap[port] = rule.Name
			}

			cb.index[proto] = pmap
		}
	}

	return nil
}

func (cb *ClusterBuilder) Match(n *node.Node) (cluster string, ok bool) {
	if len(cb.index) == 0 {
		return "", false
	}

	matches := make(map[string]int)

	for i := 0; i < len(n.Ports); i++ {
		if cluster, ok = cb.clusterFor(&n.Ports[i]); !ok {
			continue
		}

		matches[cluster]++
	}

	type match struct {
		Name   string
		Weight int
	}

	smatches := make([]*match, 0, len(matches))

	for k, n := range matches {
		w := cb.weights[k]

		smatches = append(smatches, &match{Name: k, Weight: n * w})
	}

	switch len(smatches) {
	case 0:
		return "", false
	case 1:
		return smatches[0].Name, true
	}

	// step 1: sort by rule names
	slices.SortStableFunc(smatches, func(a, b *match) int {
		return cmp.Compare(a.Name, b.Name)
	})

	// step 2: sort by weight
	slices.SortStableFunc(smatches, func(a, b *match) int {
		return cmp.Compare(a.Weight, b.Weight)
	})

	return smatches[len(smatches)-1].Name, true
}

func (cb *ClusterBuilder) clusterFor(p *node.Port) (cluster string, ok bool) {
	var pmap map[int]string

	if pmap, ok = cb.index[p.Kind]; !ok {
		return
	}

	cluster, ok = pmap[p.Value]

	return
}

func parseRulePorts(v string) (proto string, ports []int, err error) {
	const (
		protoSep = "/"
		rangeSep = "-"
		partsLen = 2
	)

	parts := strings.SplitN(v, protoSep, partsLen)
	if len(parts) != partsLen {
		return "", nil, ErrInvalidFormat
	}

	proto = parts[1]

	if strings.Contains(parts[0], rangeSep) {
		if ports, err = parsePortsRange(parts[0]); err != nil {
			return "", nil, fmt.Errorf("range: %w", err)
		}

		return proto, ports, nil
	}

	val, cerr := strconv.Atoi(parts[0])
	if cerr != nil {
		return "", nil, fmt.Errorf("port: %w", cerr)
	}

	return proto, []int{val}, nil
}

func parsePortsRange(v string) (ports []int, err error) {
	const (
		rangeSep = "-"
		partsLen = 2
	)

	prange := strings.SplitN(v, rangeSep, partsLen)

	start, err := strconv.Atoi(prange[0])
	if err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	end, err := strconv.Atoi(prange[1])
	if err != nil {
		return nil, fmt.Errorf("end: %w", err)
	}

	switch {
	case start > end:
		return nil, ErrInvalidRange
	case start == end:
		return []int{start}, nil
	}

	end++

	ports = make([]int, end-start)

	for i := start; i < end; i++ {
		ports[i-start] = i
	}

	return ports, nil
}
