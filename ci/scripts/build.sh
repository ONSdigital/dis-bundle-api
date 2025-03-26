#!/bin/bash -eux

pushd dis-bundle-api
  make build
  cp build/dis-bundle-api Dockerfile.concourse ../build
popd
