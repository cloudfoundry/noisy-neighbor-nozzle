#!/bin/bash
set -e

if [ -z $OUTPUT ]; then
    echo "OUTPUT (env variable) is required"
    exit 1
fi

if [ -z $VERSION_MAJOR ]; then
    echo "VERSION_MAJOR (env variable) is required"
    exit 1
fi

if [ -z $VERSION_MINOR ]; then
    echo "VERSION_MINOR (env variable) is required"
    exit 1
fi

if [ -z $VERSION_PATCH ]; then
    echo "VERSION_PATCH (env variable) is required"
    exit 1
fi

# Create output dir
[[ -d $OUTPUT ]] || mkdir -p $OUTPUT

# establish version/tag strings
version="{\"Major\":$VERSION_MAJOR,\"Minor\":$VERSION_MINOR,\"Build\":$VERSION_PATCH}"
semver="v$VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH"

export WORKSPACE=$PWD

# Build CF app binaries
pushd $GOPATH/src/code.cloudfoundry.org/noisy-neighbor-nozzle/
    GOOS=linux go get ./...
    for c in $(ls cmd | grep -v deployer | grep -v cli); do
        pushd cmd/$c
            GOOS=linux go build -o $OUTPUT/$c
        popd
    done
popd

# Build CLI binaries
pushd $GOPATH/src/code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/cli-plugin
    GOOS=linux go build -ldflags "-X main.version=$version" -o $OUTPUT/noisy-neighbor-cli-plugin-linux
    GOOS=darwin go build -ldflags "-X main.version=$version" -o $OUTPUT/noisy-neighbor-cli-plugin-darwin
    GOOS=windows go build -ldflags "-X main.version=$version" -o $OUTPUT/noisy-neighbor-cli-plugin-windows
popd

# Build Deployer binaries
pushd $GOPATH/src/code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/deployer
    GOOS=linux go build -o $OUTPUT/deployer-darwin
    GOOS=darwin go build -o $OUTPUT/deployer-darwin
    GOOS=windows go build -o $OUTPUT/deployer-windows
popd

# Copy manifest
cp $GOPATH/src/code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/deployer/manifest/manifest_template.yml $OUTPUT
