#!/bin/bash -eux

# Run component tests in docker compose defined in features/steps/compose folder
pushd dis-bundle-api/features/steps/compose
  COMPONENT_TEST_USE_LOG_FILE=true docker-compose up --abort-on-container-exit
  e=$?
popd

# Cat the component-test output file and remove it so log output can
# be seen in Concourse
pushd dis-bundle-api
  cat component-output.txt && rm component-output.txt
popd

# exit with the same code returned by docker compose
exit $e
