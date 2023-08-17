[![License](https://img.shields.io/badge/license-MIT%20License-blue.svg)](https://github.com/s0rg/decompose/blob/master/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/s0rg/decompose)](go.mod)
[![Release](https://img.shields.io/github/v/release/s0rg/decompose)](https://github.com/s0rg/decompose/releases/latest)

[![CI](https://github.com/s0rg/decompose/workflows/ci/badge.svg)](https://github.com/s0rg/decompose/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/s0rg/decompose)](https://goreportcard.com/report/github.com/s0rg/decompose)
[![Maintainability](https://api.codeclimate.com/v1/badges/1bc7c04689cf612a0f39/maintainability)](https://codeclimate.com/github/s0rg/decompose/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/1bc7c04689cf612a0f39/test_coverage)](https://codeclimate.com/github/s0rg/decompose/test_coverage)
![Issues](https://img.shields.io/github/issues/s0rg/decompose)

# decompose

Reverse-engineering tool for docker environments.


Takes all network connections from your docker containers, and produces [graphviz
dot](https://www.graphviz.org/doc/info/lang.html) or json stream of elements:

```
type Node struct {
    Name       string              `json:"name"`            // container name
    Image      *string             `json:"image,omitempty"` // docker image (if any)
    IsExternal bool                `json:"is_external"`     // 'external' flag - this host is from outside
    Listen     []string            `json:"listen"`          // ports description i.e. '443/tcp'
    Connected  map[string][]string `json:"connected"`       // mapping name -> ports slice
}
```


# features

- produces detailed system description with ports
- fast, it scans ~400 containers in around 5 seconds
- 100% test-coverage


# known limitations

- runs only on linux, as it uses nsenter
- runs only from root, same reason
- only established and listen connections are listed


# usage

```
decompose [flags]

possible flags with default values:

  -follow string
        follow only this container by name
  -format string
        output format: json or dot (default "dot")
  -help
        show this help
  -load value
        load json stream, can be used multiple times
  -local
        skip external hosts
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

Rebuild graph from json streams, filter by protocol, skip remote hosts and save as `dot` (no need to be root):
```
decompose -local -proto tcp -load nodes-1.json -load nodes-2.json > graph.dot
```
