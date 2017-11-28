#!/bin/sh

version="$(git describe --tags --always --dirty)"
builddate="$(date +"%Y-%m-%d")"

go build -ldflags "-X 'main.Version=$version' -X 'main.BuildDate=$builddate'"
