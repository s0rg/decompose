//go:build !test

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/graph"
)

const (
	appName       = "Decompose"
	appSite       = "https://github.com/s0rg/decompose"
	defaultProto  = "all"
	defaultFormat = "dot"
	defaultOutput = "-"
)

// build-time values.
var (
	GitTag    string
	GitHash   string
	BuildDate string
)

var (
	fSilent, fVersion bool
	fHelp, fLocal     bool
	fProto, fFormat   string
	fOut, fFollow     string
	fLoad             []string
)

func version() string {
	return fmt.Sprintf("%s %s-%s build at: %s with %s site: %s",
		appName,
		GitTag,
		GitHash,
		BuildDate,
		runtime.Version(),
		appSite,
	)
}

func usage() {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s - reverse-engineering tool for docker environments, usage:\n\n", appName)
	fmt.Fprintf(&sb, "%s [flags]\n\n", filepath.Base(os.Args[0]))
	fmt.Fprint(&sb, "possible flags with default values:\n\n")

	_, _ = os.Stderr.WriteString(sb.String())

	flag.PrintDefaults()
}

func setupFlags() {
	flag.BoolVar(&fSilent, "silent", false, "suppress progress messages in stderr")
	flag.BoolVar(&fVersion, "version", false, "show version")
	flag.BoolVar(&fHelp, "help", false, "show this help")
	flag.BoolVar(&fLocal, "local", false, "skip external hosts")

	flag.StringVar(&fProto, "proto", defaultProto, "protocol to scan: tcp, udp or all")
	flag.StringVar(&fFollow, "follow", "", "follow only this container by name")
	flag.StringVar(&fFormat, "format", defaultFormat, "output format: json or dot")
	flag.StringVar(&fOut, "out", defaultOutput, "output: filename or \"-\" for stdout")

	flag.Func("load", "load json stream, can be used multiple times", func(v string) error {
		fLoad = append(fLoad, v)

		return nil
	})

	flag.Usage = usage
}

func writeOut(name string, writer func(io.Writer)) error {
	var out io.Writer = os.Stdout

	if name != defaultOutput {
		fd, err := os.Create(name)
		if err != nil {
			return fmt.Errorf("create '%s': %w", name, err)
		}

		defer fd.Close()

		out = fd
	}

	writer(out)

	return nil
}

func run() error {
	fProto = strings.ToLower(strings.TrimSpace(fProto))
	fFormat = strings.ToLower(strings.TrimSpace(fFormat))
	fFollow = strings.TrimSpace(fFollow)

	proto, ok := graph.ParseNetProto(fProto)
	if !ok {
		fmt.Printf("unknown prototol: '%s'\n", fProto)

		return nil
	}

	bldr, ok := builder.Create(fFormat)
	if !ok {
		fmt.Printf("unknown format: '%s'\n", fFormat)

		return nil
	}

	var (
		act string
		err error
	)

	if len(fLoad) > 0 {
		act = "load"
		err = doLoad(bldr, proto, fFollow, fLocal, fLoad)
	} else {
		act = "build"
		err = doBuild(bldr, proto, fFollow, fLocal)
	}

	if err != nil {
		return fmt.Errorf("%s: %w", act, err)
	}

	if err = writeOut(strings.TrimSpace(fOut), bldr.Write); err != nil {
		return fmt.Errorf("output: %w", err)
	}

	return nil
}

func doLoad(
	b graph.Builder,
	p graph.NetProto,
	follow string,
	local bool,
	files []string,
) error {
	var proto string

	switch p {
	case graph.ALL:
	case graph.TCP, graph.UDP:
		proto = p.String()
	}

	ldr := graph.NewLoader(proto, follow, local)

	for _, fn := range files {
		fd, err := os.Open(fn)
		if err != nil {
			return fmt.Errorf("open %s: %w", fn, err)
		}

		err = ldr.LoadStream(fd)
		fd.Close()

		if err != nil {
			return fmt.Errorf("load %s: %w", fn, err)
		}
	}

	if err := ldr.Build(b); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func doBuild(
	b graph.Builder,
	p graph.NetProto,
	follow string,
	local bool,
) error {
	cli, err := graph.NewDockerClient()
	if err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	defer cli.Close()

	log.Println("Starting with method:", cli.Kind())

	if err = graph.Build(cli, b, p, follow, local); err != nil {
		return fmt.Errorf("graph: %w", err)
	}

	return nil
}

func main() {
	setupFlags()

	flag.Parse()

	if fVersion {
		fmt.Println(version())

		return
	}

	if fHelp {
		usage()

		return
	}

	if fSilent {
		log.SetOutput(io.Discard)
	}

	if err := run(); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(err)
	}
}
