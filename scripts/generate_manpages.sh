#!/usr/bin/env bash
set -euo pipefail
shopt -s extglob nullglob
IFS=$'\n\t'

version="${1:?you must set version as the first argument}"

rm -rf manpages
mkdir manpages
cd mangen
go run ./mangen.go "${version}" | gzip -c -9 > ../manpages/logalize.1.gz
