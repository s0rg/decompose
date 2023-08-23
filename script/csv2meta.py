#!/usr/bin/env python3

import argparse
import csv
import json
import sys

parser = argparse.ArgumentParser(
    description='csv2meta converter for decompose')
parser.add_argument('--skip',
                    type=int,
                    default=1,
                    help='header lines to skip (default: 1)')
parser.add_argument('incsv', type=str, help='csv for convertion')


def main():
    args = parser.parse_args()
    state = {}

    with open(args.incsv, newline='') as fd:
        for n, row in enumerate(csv.reader(fd)):
            if n < args.skip:
                continue
            key, info, tags = row[:3]
            skey = key.strip()
            if not skey:
                continue
            state[skey] = {
                'info': info.strip(),
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
