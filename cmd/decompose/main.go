//go:build !test

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/s0rg/decompose/internal/builder"
	"github.com/s0rg/decompose/internal/client"
	"github.com/s0rg/decompose/internal/graph"
)

const (
	appName       = "Decompose"
	appSite       = "https://github.com/s0rg/decompose"
	linuxOS       = "linux"
	defaultProto  = "all"
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
	fFull, fNoLoops   bool
	fProto, fFormat   string
	fOut, fFollow     string
	fMeta, fCluster   string
	fSkipEnv          string
	fLoad             []string

	ErrUnknown = errors.New("unknown")
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
	flag.BoolVar(&fFull, "full", false, "extract full process info: (cmd, args, env) and volumes info")
	flag.BoolVar(&fNoLoops, "no-loops", false, "remove connection loops (node to itself) from output")

	flag.StringVar(&fOut, "out", defaultOutput, "output: filename or \"-\" for stdout")
	flag.StringVar(&fMeta, "meta", "", "json file with metadata for enrichment")
	flag.StringVar(&fProto, "proto", defaultProto, "protocol to scan: tcp, udp or all")
	flag.StringVar(&fFollow, "follow", "", "follow only this container by name")
	flag.StringVar(&fCluster, "cluster", "", "json file with clusterization rules")
	flag.StringVar(
		&fFormat,
		"format",
		builder.KindJSON,
		"output format: json, dot, yaml, stat, tree or sdsl for structurizr dsl",
	)
	flag.StringVar(
		&fSkipEnv,
		"skip-env",
		"",
		"environment variables name(s) to skip from output, case-independent, comma-separated",
	)

	flag.Func("load", "load json stream, can be used multiple times", func(v string) error {
		res, err := filepath.Glob(v)
		if err != nil {
			return fmt.Errorf("glob '%s': %w", v, err)
		}

		fLoad = append(fLoad, res...)

		return nil
	})

	flag.Usage = usage
}

func write(name string, writer func(io.Writer) error) error {
	var out io.Writer = os.Stdout

	if name != defaultOutput {
		fd, err := os.Create(name)
		if err != nil {
			return fmt.Errorf("create '%s': %w", name, err)
		}

		defer fd.Close()

		out = fd
	}

	if err := writer(out); err != nil {
		return fmt.Errorf("write '%s': %w", name, err)
	}

	return nil
}

func feed(name string, read func(io.Reader) error) (err error) {
	fd, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("open '%s': %w", name, err)
	}

	defer fd.Close()

	if err = read(fd); err != nil {
		return fmt.Errorf("read '%s': %w", name, err)
	}

	return nil
}

func prepareConfig() (
	cfg *graph.Config,
	nwr graph.NamedWriter,
	err error,
) {
	bildr, ok := builder.Create(fFormat)
	if !ok {
		return nil, nil, fmt.Errorf(
			"%w format: %s known: %s",
			ErrUnknown,
			fFormat,
			strings.Join(builder.Names(), ","),
		)
	}

	nwr = bildr

	proto, ok := graph.ParseNetProto(fProto)
	if !ok {
		return nil, nil, fmt.Errorf("%w protocol: %s", ErrUnknown, fProto)
	}

	meta := graph.NewMetaLoader()

	if fMeta != "" {
		if err = feed(fMeta, meta.FromReader); err != nil {
			return nil, nil, fmt.Errorf("meta: %w", err)
		}
	}

	if fCluster != "" {
		if builder.SupportCluster(fFormat) {
			cluster := graph.NewClusterBuilder(bildr, nil)

			if err = feed(fCluster, cluster.FromReader); err != nil {
				return nil, nil, fmt.Errorf("cluster: %w", err)
			}

			log.Printf("Cluster rules loaded: %d", cluster.CountRules())

			bildr, nwr = cluster, cluster
		} else {
			log.Println(bildr.Name(), "cannot handle graph clusters - ignoring")
		}
	}

	skipKeys := []string{}

	if fSkipEnv != "" {
		if fFull {
			skipKeys = strings.Split(fSkipEnv, ",")
		} else {
			log.Println("skip-env makes no sense without full info - ignoring")
		}
	}

	cfg = &graph.Config{
		Builder:   bildr,
		Meta:      meta,
		Proto:     proto,
		Follow:    fFollow,
		OnlyLocal: fLocal,
		FullInfo:  fFull,
		NoLoops:   fNoLoops,
		SkipEnv:   skipKeys,
	}

	return cfg, nwr, nil
}

func run() error {
	cfg, nwr, err := prepareConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	var act string

	if len(fLoad) > 0 {
		log.Printf("Loading %d file(s)", len(fLoad))

		act, err = "load", doLoad(cfg, fLoad)
	} else {
		log.Println("Building graph")

		act, err = "build", doBuild(cfg)
	}

	if err != nil {
		return fmt.Errorf("%s: %w", act, err)
	}

	log.Println("Writing:", nwr.Name())

	if err = write(fOut, nwr.Write); err != nil {
		return fmt.Errorf("output: %w", err)
	}

	return nil
}

func doLoad(
	cfg *graph.Config,
	files []string,
) error {
	ldr := graph.NewLoader(cfg)

	for _, fn := range files {
		if err := feed(fn, ldr.FromReader); err != nil {
			return fmt.Errorf("load %s: %w", fn, err)
		}
	}

	if err := ldr.Build(); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func doBuild(
	cfg *graph.Config,
) error {
	opts := []client.Option{
		client.WithClientCreator(client.Default),
	}

	mode := client.InContainer

	if runtime.GOOS == linuxOS && os.Geteuid() == 0 {
		opts = append(opts, client.WithNsEnter(client.Nsenter))
		mode = client.LinuxNsenter
	}

	cli, err := client.NewDocker(append(opts, client.WithMode(mode))...)
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	defer cli.Close()

	log.Println("Starting with method:", cli.Mode())

	if err = graph.Build(cfg, cli); err != nil {
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
