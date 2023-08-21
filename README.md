[![License](https://img.shields.io/badge/license-MIT%20License-blue.svg)](https://github.com/s0rg/decompose/blob/master/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/s0rg/decompose)](go.mod)
[![Release](https://img.shields.io/github/v/release/s0rg/decompose)](https://github.com/s0rg/decompose/releases/latest)
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
- json stream of items:
```go
type Item struct {
    Name       string              `json:"name"`            // container name
    Image      *string             `json:"image,omitempty"` // docker image (if any)
    IsExternal bool                `json:"is_external"`     // this host is external
    Meta       *Meta               `json:"meta,omitempty"`  // metadata, see below
    Listen     []string            `json:"listen"`          // ports description i.e. '443/tcp'
    Networks   []string            `json:"networks"`        // network names
    Connected  map[string][]string `json:"connected"`       // name -> ports slice
}

type Meta struct {
    Info string   `json:"info"`
    Tags []string `json:"tags"`
}
```

example:

```json
{
    "name": "foo-1",
    "image": "repo/foo:latest",
    "is_external": false,
    "meta": {
        "info": "foo info",
        "tags": ["foo"]
    },
    "listen": ["80/tcp"],
    "networks": ["test-net"],
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
one of provided keys, like `foo-1` or `bar-1` for this example.


# features

- os-independent, it uses different strategies to get container connections:
    * running on **linux as root** is the fastest way and it will work with all types of containers (even
            `scratch`-based)
    * running as non-root or on non-linux OS will attempt to run `netsat` inside container, if this fails
    (i.e. for missing `netstat` binary), no connections for such container will be gathered
- produces detailed connections graph with ports
- fast, it scans ~400 containers in around 5 seconds
- 100% test-coverage


# known limitations

- only established and listen connections are listed


# usage

```
decompose [flags]

possible flags with default values:

  -follow string
        follow only this container by name
  -format string
        output format: json, dot or sdsl for structurizr dsl (default "dot")
  -help
        show this help
  -load value
        load json stream, can be used multiple times
  -local
        skip external hosts
  -meta string
        json with metadata (info and tags) to enrich output graph
  -out string
        output: filename or "-" for stdout (default "-")
  -proto string
        protocol to scan: tcp, udp or all (default "all")
  -silent
        suppress progress messages in stderr
  -version
        show version
```


# examples

Get `dot` file:
```
sudo decompose > connections.dot
```

Get json stream:
```
sudo decompose -format json | jq '{name}'
```

Get only tcp connections as `dot`:
```
sudo decompose -proto tcp > tcp.dot
```

Save json stream:
```
sudo decompose -format json > nodes-1.json
```

Merge graphs from json streams, filter by protocol, skip remote hosts and save as `dot` (no need to be root):
```
decompose -local -proto tcp -load nodes-1.json -load nodes-2.json > graph-merged.dot
```


# example result

Scheme taken from [redis-cluster](https://github.com/s0rg/redis-cluster-compose):


![svg](https://github.com/s0rg/redis-cluster-compose/blob/main/redis-cluster.svg)


Steps to reproduce:

```shell
git clone https://github.com/s0rg/redis-cluster-compose.git
cd redis-cluster-compose
docker compose up
```

in other terminal:

```shell
sudo decompose | dot -Tsvg > redis-cluster.svg
```
