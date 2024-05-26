#!/usr/bin/env bash
set -euo pipefail
shopt -s extglob nullglob
IFS=$'\n\t'

rm -rf completions
mkdir completions
cd compgen
for shell in bash zsh fish; do
	go run ./compgen.go "${shell}" > "../completions/logalize.${shell}"
done
