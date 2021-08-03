#!/bin/bash

set -e

CONTROLLER_IMG ?= kubedl/morphling-controllers:latest
DB_MANAGER_IMG ?= kubedl/morphling-database-manager:latest
UI_IMG ?= kubedl/morphling-ui:latest
ALGORITHM_IMG ?= kubedl/morphling-algorithm:base
CLIENT_IMG ?= kubedl/morphling-http-client:demo

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/../..
cd ${SCRIPT_ROOT}
echo "cd to ${SCRIPT_ROOT}"

# controller, storage, and ui
docker build -t ${UI_IMG} -f cmd/ui/Dockerfile .
docker build -t ${DB_MANAGER_IMG} -f cmd/db-manager/Dockerfile .
docker build -t ${CONTROLLER_IMG} -f cmd/controllers/Dockerfile .

# algorithm server
docker build -t ${ALGORITHM_IMG} -f cmd/algorithm/grid/Dockerfile .

# http client
cp api/v1alpha1/grpc_proto/grpc_storage/python3/* pkg/client/
cd pkg/client/
docker build -t ${CLIENT_IMG} -f ./Dockerfile .

echo -e "\n Docker images build succeeded\n"
