#!/usr/bin/env bash

#set -e

set -o xtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

cd ${SCRIPT_ROOT}
kustomize build manifests/ | kubectl apply -f -