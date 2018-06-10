#!/bin/bash

outdir=release
if [ ! -d "$outdir" ]; then
    mkdir "$outdir"
fi

VERSION=main.Version=`git describe --tags`
BUILT=main.Built=`date +%FT%T%z`
GIT_COMMIT=main.GitCommit=`git rev-parse --short HEAD`
GO_VERSION=main.GoVersion=`go version`
ldflags="-X '$VERSION' -X '$BUILT' -X '$GIT_COMMIT' -X '$GO_VERSION'"

export GOOS=windows
export GOARCH=amd64
echo building: $GOOS $GOARCH
go build -o "$outdir/zworld.exe" -ldflags "$ldflags" zworld.go

export GOOS=windows
export GOARCH=386
echo building: $GOOS $GOARCH
go build -o "$outdir/zworld-x32.exe" -ldflags "$ldflags" zworld.go

export GOOS=linux
export GOARCH=amd64
echo building: $GOOS $GOARCH
go build -o "$outdir/zworld" -ldflags "$ldflags" zworld.go
