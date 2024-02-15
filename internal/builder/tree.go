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

	fmt.Fprint(w, next, " ")
	// fmt.Fprintln(w, "listen:", strings.Join(n.Listen, ", "))

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

	slices.SortStableFunc(dstOrder, cmp.Compare)

	for _, dst := range dstOrder {
		ports := n.Connected[dst]

		fmt.Fprint(w, next, " ")

		if cur == lst {
			fmt.Fprint(w, symEnd)
		} else {
			fmt.Fprint(w, symEdge)
		}

		_ = ports
		// fmt.Fprintf(w, " %s: %s\n", dst, strings.Join(ports, ", "))

		cur++
	}

	if !last {
		fmt.Fprintln(w, next)
	}
}
