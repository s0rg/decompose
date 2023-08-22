package srtructurizr

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"unicode"
)

const tab = "\t"

func putKey(
	w io.Writer,
	level int,
	key, value string,
) {
	if value == "" {
		return
	}

	fmt.Fprint(w, strings.Repeat(tab, level))
	fmt.Fprintf(w, `%s "%s"`, key, value)
	fmt.Fprintln(w, "")
}

func putView(
	w io.Writer,
	level int,
	key, id string,
) {
	fmt.Fprint(w, strings.Repeat(tab, level))
	fmt.Fprintf(w, `%s %s "%s" {`, key, id, id)
	fmt.Fprintln(w, "")
}

func putRaw(
	w io.Writer,
	level int,
	raw string,
) {
	fmt.Fprint(w, strings.Repeat(tab, level))
	fmt.Fprintf(w, "%s\n", raw)
}

func putHeader(
	w io.Writer,
	level int,
	key string,
) {
	fmt.Fprint(w, strings.Repeat(tab, level))
	fmt.Fprintf(w, "%s {\n", key)
}

func putCommon(
	w io.Writer,
	level int,
	desc, tech string,
	tags []string,
) {
	putKey(w, level, keyDescription, desc)
	putKey(w, level, keyTechnology, tech)

	if ctags, ok := compactTags(tags); ok {
		putKey(w, level, keyTags, strings.Join(ctags, ","))
	}
}

func putBlock(
	w io.Writer,
	level int,
	block, key, value string,
) {
	fmt.Fprint(w, strings.Repeat(tab, level))
	fmt.Fprintf(w, `%s = %s "%s" {`, key, block, value)
	fmt.Fprintln(w, "")
}

func putRelation(
	w io.Writer,
	level int,
	src, dst string,
) {
	fmt.Fprint(w, strings.Repeat(tab, level))
	fmt.Fprintf(w, "%s -> %s {\n", src, dst)
}

func putEnd(w io.Writer, level int) {
	fmt.Fprint(w, strings.Repeat(tab, level))
	fmt.Fprintln(w, "}")
}

func safeID(v string) (id string) {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsSpace(r), r == '-', r == '.', r == ':':
			return '_'
		}

		return r
	}, v)
}

func compactTags(tags []string) (rv []string, ok bool) {
	if len(tags) == 0 {
		return nil, false
	}

	rv = slices.DeleteFunc(tags, func(v string) bool {
		return strings.TrimSpace(v) == ""
	})

	switch len(rv) {
	case 0:
		return nil, false
	case 1:
		return rv, true
	}

	sort.Strings(rv)

	rv = slices.Clip(slices.Compact(rv))

	return rv, true
}
