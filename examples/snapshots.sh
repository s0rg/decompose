#!/usr/bin/env bash

# working directory for snapshots
OUT=snapshots
# number of snapshots to take
COUNT=5
# wait time (in seconds) between each snapshot
WAIT=10

mkdir -p "${OUT}"

for i in $(seq 1 ${COUNT}); do
	echo "Taking snapshot ${i}..."

	decompose -format json -out "${OUT}/snapshot_${i}.json"

	if [[ "${i}" -ne "${COUNT}" ]]; then
		echo "Sleeping for ${WAIT} seconds..."
		sleep "${WAIT}"
	fi
done

echo "Merging..."

bin/decompose -load "${OUT}/*.json" -format json -out merged.json

echo "Cleaning-up..."

rm -rf "${OUT}"
