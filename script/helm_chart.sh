#!/bin/bash

set -e

CONTROLLER_IMG=kubedl/morphling-controllers:latest
DB_MANAGER_IMG=kubedl/morphling-database-manager:latest
UI_IMG=kubedl/morphling-ui:latest
ALGORITHM_IMG=kubedl/morphling-algorithm:base
CLIENT_IMG=kubedl/morphling-http-client:demo

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
cd ${SCRIPT_ROOT}
echo "cd to ${SCRIPT_ROOT}"

# copy files and replace context
cp config/crd/bases/* helm/morphling/crds
cp config/rbac/role.yaml helm/morphling/templates
sed -i.morphlingbackup 's/name:.*/name: {{ include "morphling.fullname" . }}-role/g' helm/morphling/templates/role.yaml
cp -r manifests/* helm/morphling/templates
rm -f helm/morphling/*.morphlingbackup
rm -f helm/morphling/templates/*.morphlingbackup
#rm -f helm/morphling/kustomization.yaml
find helm/morphling/templates -type f -name 'kustomization.yaml' -exec rm {} +