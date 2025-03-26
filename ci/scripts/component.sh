#!/bin/bash -eux

pushd dis-bundle-api
  make test-component
popd
