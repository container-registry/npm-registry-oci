#!/usr/bin/env sh
# Do - The Simplest Build Tool on Earth.
# Documentation and examples see https://github.com/8gears/do

set -e -u # -e "Automatic exit from bash shell script on error"  -u "Treat unset variables and parameters as errors"

build() {
  CGO_ENABLED=0 go build -o .bin/registry .
}

run() {
 USE_TLS=true DEBUG=true .bin/registry registry serve
}

tests() {
  go test -v ./...
}


"$@" # <- execute the task

[ "$#" -gt 0 ] || printf "Usage:\n\t./do.sh %s\n" "($(compgen -A function | grep '^[^_]' | paste -sd '|' -))"
