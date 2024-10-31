#!/bin/bash

set -e

CONTROLLER_IMG=kubedl/morphling-controllers:latest
DB_MANAGER_IMG=kubedl/morphling-database-manager:latest
UI_IMG=kubedl/morphling-ui:latest
ALGORITHM_IMG=kubedl/morphling-algorithm:base
HTTP_CLIENT_IMG=kubedl/morphling-http-client:demo
GRPC_CLIENT_IMG=kubedl/morphling-grpc-client:demo
SERVER_IMG=kubedl/morphling-grpc-server:latest

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
cd ${SCRIPT_ROOT}
echo "cd to ${SCRIPT_ROOT}"

# controller, storage, and ui
docker build -t ${UI_IMG} -f  console/Dockerfile .
docker build -t ${DB_MANAGER_IMG} -f cmd/db-manager/Dockerfile .
docker build -t ${CONTROLLER_IMG} -f cmd/controllers/Dockerfile .

# algorithm server
docker build -t ${ALGORITHM_IMG} -f cmd/algorithm/grid/Dockerfile .

# http client
cp api/v1alpha1/grpc_proto/grpc_storage/python3/* pkg/client/
cd pkg/client/
docker build -t ${HTTP_CLIENT_IMG} -f ./Dockerfile .

# grpc client
cd -
cp api/v1alpha1/grpc_proto/grpc_storage_v2/python3/* pkg/grpc_client/
cp api/v1alpha1/grpc_proto/grpc_predict/python3/* pkg/grpc_client/
cd pkg/grpc_client/
docker build -t ${GRPC_CLIENT_IMG} -f ./Dockerfile .

# multi inference framework grpc server
cd -
CUDA_VERSION=118
VLLM_FILE=vllm-0.6.1.post1+cu118-cp310-cp310-manylinux1_x86_64.whl
cp api/v1alpha1/grpc_proto/grpc_predict/python3/* pkg/server/
cd pkg/server
docker build --build-arg CUDA_VERSION=${CUDA_VERSION} --build-arg VLLM_FILE=${VLLM_FILE} -t ${SERVER_IMG} -f Dockerfile .

echo -e "\n Docker images build succeeded\n"
