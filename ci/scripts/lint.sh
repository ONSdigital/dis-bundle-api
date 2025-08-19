#!/bin/bash -eux

go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

pushd dis-bundle-api
  make lint
popd
