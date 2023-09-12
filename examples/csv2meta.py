#!/usr/bin/env python3

import argparse
import csv
import json
import sys

parser = argparse.ArgumentParser(
    description='csv2meta converter for decompose')
parser.add_argument(
    '--skip',
    type=int,
    default=1,
    help='header lines to skip (default: 1)',
)
parser.add_argument(
    'csv',
    type=str,
    help=
    'csv for convertion, with 5 columns (at least): key, info, docs, repo, tags'
)


def main():
    args = parser.parse_args()
    state = {}

    with open(args.csv, newline='') as fd:
        for n, row in enumerate(csv.reader(fd)):
            if n < args.skip:
                continue
            key, info, docs, repo, tags = row[:5]
            skey = key.strip()
            if not skey:
                continue
            state[skey] = {
                'info': info.strip(),
                'docs': docs.strip(),
                'repo': repo.strip(),
                'tags': list(filter(bool, tags.split(','))),
            }

    json.dump(
        state,
        sys.stdout,
        sort_keys=True,
        indent=4,
        ensure_ascii=False,
    )


main()
