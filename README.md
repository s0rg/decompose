[![License](https://img.shields.io/badge/license-MIT%20License-blue.svg)](https://github.com/s0rg/decompose/blob/master/LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose?ref=badge_shield)
[![Go Version](https://img.shields.io/github/go-mod/go-version/s0rg/decompose)](go.mod)
[![Release](https://img.shields.io/github/v/release/s0rg/decompose)](https://github.com/s0rg/decompose/releases/latest)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
![Downloads](https://img.shields.io/github/downloads/s0rg/decompose/total.svg)

[![CI](https://github.com/s0rg/decompose/workflows/ci/badge.svg)](https://github.com/s0rg/decompose/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/s0rg/decompose)](https://goreportcard.com/report/github.com/s0rg/decompose)
[![Maintainability](https://api.codeclimate.com/v1/badges/1bc7c04689cf612a0f39/maintainability)](https://codeclimate.com/github/s0rg/decompose/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/1bc7c04689cf612a0f39/test_coverage)](https://codeclimate.com/github/s0rg/decompose/test_coverage)
![Issues](https://img.shields.io/github/issues/s0rg/decompose)

# decompose

Reverse-engineering tool for docker environments.

Takes all network connections from your docker containers and exports them as:

- [graphviz dot](https://www.graphviz.org/doc/info/lang.html)
- [structurizr dsl](https://github.com/structurizr/dsl)
- [compose yaml](https://github.com/compose-spec/compose-spec/blob/master/spec.md)
- pseudographical tree
- json stream
- statistics - nodes, connections and listen ports counts

## rationale

I was in need for a tool to visualize and inspect big (more than 470 containers) dockerized legacy system without any
schemes and having a bare minimum of documentation

## analogs

Closest analogs, i can find, that not suit my needs very well:

- [Red5d/docker-autocompose](https://github.com/Red5d/docker-autocompose) - produces only `compose yaml`
- [justone/dockviz](https://github.com/justone/dockviz) - produces only `dot`, links and ports are taken
  from compose configuration (`links` and `ports` sections) directly, therefore can miss some of them
- [LeoVerto/docker-network-graph](https://github.com/LeoVerto/docker-network-graph) - very same as above, in python
- [weaveworks/scope](https://github.com/weaveworks/scope) - deprecated, no cli

## features

- os-independent, it uses different strategies to get container connections:
  - running on **linux as root** is the fastest way and it will work with all types of containers (even `scratch`-based)
    as it use `nsenter`
  - running as non-root or on non-linux OS will attempt to run `netsat` inside container, if this fails
    (i.e. for missing `netstat` binary), no connections for such container will be gathered
- single-binary, static-compiled unix-way `cli` (all output goes to stdout, progress information to stderr)
- produces detailed connections graph **with ports**
- save `json` stream once and process it later in any way you want
- all output formats are sorted, thus can be placed to any `vcs` to observe changes
- fast, scans ~470 containers with ~4000 connections in around 5 sec
- auto-clusterization based on graph topology
- 100% test-coverage

## known limitations

- only established and listen connections are listed (but script like [snapshots.sh](examples/snapshots.sh) can beat this)
- `composer-yaml` is not intended to be working out from the box, it can lack some of crucial information (even in `-full` mode),
  or may contains cycles between nodes (removing `links` section in services may help), its main purpose is for system overview

## installation

- [binaries / deb / rpm](https://github.com/s0rg/decompose/releases) for Linux, FreeBSD, macOS and Windows
- [docker image](https://hub.docker.com/r/s0rg/decompose)

## usage

```
decompose [flags]

possible flags with default values:

  -cluster string
        json file with clusterization rules, or auto:<similarity> for auto-clustering, similarity is float in (0.0, 1.0] range
  -follow string
        follow only this container by name(s), comma-separated or from @file
  -format string
        output format: json, dot, yaml, stat, tree or sdsl for structurizr dsl (default "json")
  -full
        extract full process info: (cmd, args, env) and volumes info
  -help
        show this help
  -load value
        load json stream, can be used multiple times
  -local
        skip external hosts
  -meta string
        json file with metadata for enrichment
  -no-loops
        remove connection loops (node to itself) from output
  -out string
        output: filename or "-" for stdout (default "-")
  -proto string
        protocol to scan: tcp, udp or all (default "all")
  -silent
        suppress progress messages in stderr
  -skip-env string
        environment variables name(s) to skip from output, case-independent, comma-separated
  -version
        show version
```

### environment variables:

- `DOCKER_HOST` - connection uri
- `DOCKER_CERT_PATH` - directory path containing key.pem, cert.pm and ca.pem
- `DOCKER_TLS_VERIFY` - enable client TLS verification
- `IN_DOCKER_PROC_ROOT` - for in-docker scenario - root for host-mounted /proc

## json stream format

```go
type Item struct {
    Name       string              `json:"name"` // container name
    IsExternal bool                `json:"is_external"` // this host is external
    Image      *string             `json:"image,omitempty"` // docker image (if any)
    Process    *struct{
        Cmd []string `json:"cmd"`
        Env []string `json:"env"`
    } `json:"process,omitempty"` // process info, only when '-full'
    Listen     []string            `json:"listen"` // ports description i.e. '443/tcp'
    Networks   []string            `json:"networks"` // network names
    Tags       []string            `json:"tags"` // tags, if meta presents
    Volumes    []*struct{
        Type string `json:"type"`
        Src  string `json:"src"`
        Dst  string `json:"dst"`
    } `json:"volumes"`           // volumes info, only when '-full'
    Connected  map[string][]string `json:"connected"` // name -> ports slice
}
```

Single node example with full info and metadata filled:

```json
{
    "name": "foo-1",
    "is_external": false,
    "image": "repo/foo:latest",
    "process": {
        "cmd": [
            "foo",
            "-foo-arg"
        ],
        "env": [
            "FOO=1"
        ]
    },
    "listen": ["80/tcp"],
    "networks": ["test-net"],
    "tags": ["some"],
    "volumes": [
        {
            "type": "volume",
            "src": "/var/lib/docker/volumes/foo_1/_data",
            "dst": "/data"
        },
        {
            "type": "bind",
            "src": "/path/to/foo.conf",
            "dst": "/etc/foo.conf"
        }
    ],
    "connected": {
        "bar-1": ["443/tcp"]
    }
}
```

See [stream.json](examples/stream.json) for simple stream example.

## metadata format

To enrich output with detailed descriptions, you can provide additional `json` file, with metadata i.e.:

```json
{
    "foo": {
        "info": "info for foo",
        "docs": "https://acme.corp/docs/foo",
        "repo": "https://git.acme.corp/foo",
        "tags": ["some"]
    },
    "bar": {
        "info": "info for bar",
        "tags": ["other", "not-foo"]
    }
}
```

Using this file `decompose` can enrich output with info and additional tags, for every container that match by name with
one of provided keys, like `foo-1` or `bar1` for this example.

See [csv2meta.py](examples/csv2meta.py) for example how to create such `json` fom csv, and
[meta.json](examples/meta.json) for metadata sample.

## clusterization

### with rules

You can join your services into `clusters` by flexible rules, in `dot`, `structurizr` and `stat` output formats.
Example `json` (order matters):

```json
[
    {
        "name": "cluster-name",
        "weight": 1,
        "if": "<expression>"
    },
    ...
]
```

Weight can be omitted, if not specified it equals `1`.

Where `<expression>` is [expr dsl](https://expr-lang.org/docs/Language-Definition), having env object `node` with follownig
fields:

```go
type Node struct {
    Listen     PortMatcher  // port matcher with two methods: `HasAny(...string) bool` and `Has(...string) bool`
    Name       string       // container name
    Image      string       // container image
    Cmd        string       // container cmd
    Args       []string     // container args
    Tags       []string     // tags, if meta present
    IsExternal bool         // external flag
}
```

See: [cluster.json](examples/cluster.json) for detailed example.

### automatic

Decompose provides automatic clusterization option, use `-cluster auto:<similarity>` to try it out, `similarity` is
a float in `(0.0, 1.0]` range, representing how much similar ports nodes must have to be placed in same cluster
(`1.0` - must have all ports equal).

## examples

Save full json stream:

```shell
decompose -full > nodes-1.json
```

Get `dot` file:

```shell
decompose -format dot > connections.dot
```

Get only tcp connections as `dot`:

```shell
decompose -proto tcp -format dot > tcp.dot
```

Merge graphs from json streams, filter by protocol, skip remote hosts and save as `dot`:

```shell
decompose -local -proto tcp -load "nodes-*.json" -format dot > graph-merged.dot
```

Load json stream, enrich and save as `structurizr dsl`:

```shell
decompose -load nodes-1.json -meta metadata.json -format sdsl > workspace.dsl
```

Save auto-clustered graph, with similarity factor `0.6` as `structurizr dsl`:

```shell
decompose -cluster auto:0.6 -format sdsl > workspace.dsl
```

## example result

Scheme taken from [redis-cluster](https://github.com/s0rg/redis-cluster-compose):

![svg](https://github.com/s0rg/redis-cluster-compose/blob/main/redis-cluster.svg) *it may be too heavy to display it with
browser, use `save image as` and open it locally*

Steps to reproduce:

```shell
git clone https://github.com/s0rg/redis-cluster-compose.git
cd redis-cluster-compose
docker compose up -d
```

then:

```shell
decompose -format dot | dot -Tsvg > redis-cluster.svg
```

## license

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose?ref=badge_large)
