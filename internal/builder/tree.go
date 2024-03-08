package builder

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/s0rg/decompose/internal/node"
)

const (
	space   = "  "
	symRoot = ". "
	symEdge = "├─"
	symLine = "│ "
	symEnd  = "└─"
)

type Tree struct {
	j *JSON
}

func NewTree() *Tree {
	return &Tree{
		j: NewJSON(),
	}
}

func (t *Tree) Name() string {
	return "text-tree"
}

func (t *Tree) AddNode(n *node.Node) error {
	return t.j.AddNode(n)
}

func (t *Tree) AddEdge(e *node.Edge) {
	t.j.AddEdge(e)
}

func (t *Tree) Write(w io.Writer) error {
	fmt.Fprintln(w, symRoot)

	t.j.Sorted(func(n *node.JSON, last bool) {
		writeNode(w, n, last)
	})

	return nil
}

func writeNode(w io.Writer, n *node.JSON, last bool) {
	var next string

	if last {
		fmt.Fprint(w, symEnd)

		next = space
	} else {
		fmt.Fprint(w, symEdge)

		next = symLine
	}

	fmt.Fprint(w, " ")
	fmt.Fprintln(w, n.Name)

	fmt.Fprint(w, next, " ")
	fmt.Fprintf(w, "external: %t\n", n.IsExternal)

	if n.Image != nil {
		fmt.Fprint(w, next, " ")
		fmt.Fprintln(w, "image:", *n.Image)
	}

	if len(n.Tags) > 0 {
		fmt.Fprint(w, next, " ")
		fmt.Fprintln(w, "tags:", strings.Join(n.Tags, ", "))
	}

	if len(n.Container.Cmd) > 0 {
		fmt.Fprint(w, next, " ")
		fmt.Fprintf(w, "cmd: '%s'\n", strings.Join(n.Container.Cmd, " "))
	}

	fmt.Fprint(w, next, " ")
	fmt.Fprintln(w, "listen:", joinListeners(n.Listen, ", "))

	if len(n.Networks) > 0 {
		fmt.Fprint(w, next, " ")
		fmt.Fprintln(w, "networks:", strings.Join(n.Networks, ", "))
	}

	var (
		cur int
		lst = len(n.Connected) - 1
	)

	if len(n.Connected) > 0 {
		fmt.Fprint(w, next, " ", symLine, "\n")
	}

	dstOrder := make([]string, 0, len(n.Connected))

	for dst := range n.Connected {
		dstOrder = append(dstOrder, dst)
	}

	slices.SortFunc(dstOrder, cmp.Compare)

	for _, dst := range dstOrder {
		ports := n.Connected[dst]

		fmt.Fprint(w, next, " ")

		if cur == lst {
			fmt.Fprint(w, symEnd)
		} else {
			fmt.Fprint(w, symEdge)
		}

		fmt.Fprintf(w, " %s: %s\n", dst, joinConnections(ports, ", "))

		cur++
	}

	if !last {
		fmt.Fprintln(w, next)
	}
}

func joinConnections(conns []*node.Connection, sep string) (rv string) {
	raw := make([]string, 0, len(conns))

	for _, c := range conns {
		raw = append(raw, c.Port.Label())
	}

	slices.Sort(raw)

	return strings.Join(raw, sep)
}

func joinListeners(ports map[string][]*node.Port, sep string) (rv string) {
	var tmp []string

	for _, plist := range ports {
		for _, p := range plist {
			tmp = append(tmp, p.Label())
		}
	}

	slices.Sort(tmp)

	return strings.Join(tmp, sep)
}
