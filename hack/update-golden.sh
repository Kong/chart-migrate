#!/usr/bin/env bash

# Locations in this script are from the repository root. This script
# must be run via "make test.golden.update" rather than invoked from /hack

set -o nounset
set -o pipefail

SOURCE="./test/integration/testdata/source/"
EXPECTED="./test/integration/testdata/expected/"
KONGS=("${SOURCE}"*_kong_values.yaml)
for KONG in "${KONGS[@]}";
do FILENAME="${KONG##*/}"
go run ./pkg/cmd -f "${KONG}" -s kong migrate > "${EXPECTED}${FILENAME}"
done

TMPDIR="$(mktemp -d )"
INGRESSES=("${SOURCE}"*_ingress_values.yaml)
for INGRESS in "${INGRESSES[@]}";
do FILENAME="${INGRESS##*/}"
go run ./pkg/cmd -f "${INGRESS}" -s ingress migrate > "${TMPDIR}/${FILENAME}"
go run ./pkg/cmd -f "${TMPDIR}/${FILENAME}" -s ingress merge > "${EXPECTED}${FILENAME}"
done
