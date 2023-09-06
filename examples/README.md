# examples

## example json files

- `stream.json` - simple system as json stream example
- `cluster.json` - clusterization rules example
- `meta.json` - metadata example

usage:

```shell
decompose -cluster cluster.json -meta meta.json -load stream.json -format dot | dot -Tsvg > example.svg
```

## csv2meta script

example script to convert any compatible csv - at least 3 columns, with `id, info, tags` order and comma as delimeter
to metadata json for decomposer

## usage

```shell
python3 csv2meta.py my_meta_utf8.csv > meta.json
```

## snapshots.sh script

example script for taking and merging snapshots, writes result to `merged.json`
