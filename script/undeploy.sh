#!/usr/bin/env bash

#set -e

set -o xtrace

# Delete all pe in all namespaces.
kubectl delete pe --all --all-namespaces
sleep 10

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

cd ${SCRIPT_ROOT}
kustomize build manifests/ | kubectl delete -f -
kubectl delete namespace morphling-system