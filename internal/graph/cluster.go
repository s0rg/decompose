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
		Name  string   `json:"name"`
		Ports []string `json:"ports"`
	}

	ClusterAssigner struct {
		index map[string]map[int]string
	}
)

func NewClusterAssigner() *ClusterAssigner {
	return &ClusterAssigner{
		index: make(map[string]map[int]string),
	}
}

func (ca *ClusterAssigner) IsEmpty() bool {
	return len(ca.index) == 0
}

func (ca *ClusterAssigner) FromReader(r io.Reader) (err error) {
	var rules []ruleJSON

	dec := json.NewDecoder(r)

	for dec.More() {
		if err = dec.Decode(&rules); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}

	for i := 0; i < len(rules); i++ {
		rule := &rules[i]

		for j := 0; j < len(rule.Ports); j++ {
			proto, ports, perr := parseRulePorts(rule.Ports[j])
			if perr != nil {
				return fmt.Errorf("parse '%s': %w", rule.Ports[j], perr)
			}

			pmap, ok := ca.index[proto]
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

			ca.index[proto] = pmap
		}
	}

	return nil
}

func (ca *ClusterAssigner) Match(n *node.Node) (cluster string, ok bool) {
	if ca.IsEmpty() {
		return "", false
	}

	matches := make(map[string]int)

	for i := 0; i < len(n.Ports); i++ {
		if cluster, ok = ca.clusterFor(&n.Ports[i]); !ok {
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
		smatches = append(smatches, &match{Name: k, Weight: n})
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

func (ca *ClusterAssigner) RefineConns(
	nodes map[string]*node.Node,
) (rv map[string]node.Ports, ok bool) {
	if ca.IsEmpty() {
		return nil, false
	}

	return rv, true
}

func (ca *ClusterAssigner) clusterFor(p *node.Port) (cluster string, ok bool) {
	var pmap map[int]string

	if pmap, ok = ca.index[p.Kind]; !ok {
		return
	}

	cluster, ok = pmap[p.Value]

	return
}

func (ca *ClusterAssigner) Assign(n *node.Node) {
	if cluster, ok := ca.Match(n); ok {
		n.Cluster = cluster
	}
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
		return "", nil, fmt.Errorf("atoi: %w", cerr)
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
		return nil, fmt.Errorf("atoi start: %w", err)
	}

	end, err := strconv.Atoi(prange[1])
	if err != nil {
		return nil, fmt.Errorf("atoi end: %w", err)
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
