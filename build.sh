#!/bin/bash

set -e

OS="darwin linux windows"
ARCH="amd64"

echo "Getting build dependencies"
godep get
go get -u github.com/golang/lint/golint

echo "Ensuring code quality"
go vet ./...
golint ./...

ver=$(cd $GOPATH/src/github.com/hashicorp/terraform && git describe --abbrev=0 --tags)
echo "VERSION terraform '$ver'"

for GOOS in $OS; do
    for GOARCH in $ARCH; do
        arch="$GOOS-$GOARCH"
        binary="bin/terraform-provider-clc.$arch"
        echo "Building $binary"
        GOOS=$GOOS GOARCH=$GOARCH go build -o $binary bin/terraform-provider-clc/main.go
    done
done
