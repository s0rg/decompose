[![License](https://img.shields.io/badge/license-MIT%20License-blue.svg)](https://github.com/s0rg/decompose/blob/master/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/s0rg/decompose)](go.mod)
[![Release](https://img.shields.io/github/v/release/s0rg/decompose)](https://github.com/s0rg/decompose/releases/latest)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose?ref=badge_shield)

<!-- ![Downloads](https://img.shields.io/github/downloads/s0rg/decompose/total.svg) -->

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
- pseudographical tree
- json stream of items:

```go
type Item struct {
    Name       string              `json:"name"`              // container name
    IsExternal bool                `json:"is_external"`       // this host is external
    Image      *string             `json:"image,omitempty"`   // docker image (if any)
    Meta       *Meta               `json:"meta,omitempty"`    // metadata, see below
    Process    *Process            `json:"process,omitempty"` // process info
    Listen     []string            `json:"listen"`            // ports description i.e. '443/tcp'
    Networks   []string            `json:"networks"`          // network names
    Volumes    []*Volume           `json:"volumes"`           // volumes
    Connected  map[string][]string `json:"connected"`         // name -> ports slice
}

type Meta struct {
    Info string   `json:"info"`
    Tags []string `json:"tags"`
}

type Volume struct {
    Type string `json:"type"`
    Src  string `json:"src"`
    Dst  string `json:"dst"`
}

type Process struct {
    Cmd []string `json:"cmd"`
    Env []string `json:"env"`
}
```

example with full info and metadata filled:

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
    "meta": {
        "info": "info for foo",
        "tags": ["some"]
    },
    "listen": ["80/tcp"],
    "networks": ["test-net"],
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

# metadata format

To enrich output with detailed descriptions, you can provide additional `json` file, with metadata i.e.:

```json
{
    "foo": {
        "info": "info for foo",
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

See [csv2meta.py](script/csv2meta.py) for example how to create such `json` fom csv.

# features

- os-independent, it uses different strategies to get container connections:
  - running on **linux as root** is the fastest way and it will work with all types of containers (even
    `scratch`-based)
  - running as non-root or on non-linux OS will attempt to run `netsat` inside container, if this fails
    (i.e. for missing `netstat` binary), no connections for such container will be gathered
- produces detailed connections graph **with ports**
- fast, it scans ~400 containers in around 5 seconds
- 100% test-coverage

# known limitations

- only established and listen connections are listed

# installation

- [binaries / deb / rpm](https://github.com/s0rg/decompose/releases) for Linux, FreeBSD, macOS and Windows.

# usage

```
decompose [flags]

possible flags with default values:

  -follow string
        follow only this container by name
  -format string
        output format: dot, json, tree or sdsl for structurizr dsl (default "dot")
  -full
        extract full process info: (cmd, args, env) and volumes info
  -help
        show this help
  -load value
        load json stream, can be used multiple times
  -local
        skip external hosts
  -meta string
        filename with json metadata for enrichment
  -no-loops
        remove connection loops (node to itself) from output
  -out string
        output: filename or "-" for stdout (default "-")
  -proto string
        protocol to scan: tcp, udp or all (default "all")
  -silent
        suppress progress messages in stderr
  -version
        show version
```

## environment variables:

- `DOCKER_HOST` - connection uri
- `DOCKER_CERT_PATH` - directory path containing key.pem, cert.pm and ca.pem
- `DOCKER_TLS_VERIFY` - enable client TLS verification

# examples

Get `dot` file:

```shell
decompose > connections.dot
```

Process json stream:

```shell
decompose -format json | jq '{name}'
```

Get only tcp connections as `dot`:

```shell
decompose -proto tcp > tcp.dot
```

Save full json stream:

```shell
decompose -full -format json > nodes-1.json
```

Display tree:

```shell
decompose -format tree
```

Merge graphs from json streams, filter by protocol, skip remote hosts and save as `dot`:

```shell
decompose -local -proto tcp -load nodes-1.json -load nodes-2.json > graph-merged.dot
```

Load json stream, enrich and save as `structurizr dsl`:

```shell
decompose -load nodes-1.json -meta metadata.json -format sdsl > workspace.dsl
```

# example result

Scheme taken from [redis-cluster](https://github.com/s0rg/redis-cluster-compose):

![svg](https://github.com/s0rg/redis-cluster-compose/blob/main/redis-cluster.svg) *it may be too heavy to display it with
browser, use `save image as` and open it locally*

Steps to reproduce:

```shell
git clone https://github.com/s0rg/redis-cluster-compose.git
cd redis-cluster-compose
docker compose up
```

in other terminal:

```shell
decompose | dot -Tsvg > redis-cluster.svg
```


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fs0rg%2Fdecompose?ref=badge_large)