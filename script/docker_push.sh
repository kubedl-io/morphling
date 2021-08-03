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

# controller, storage, and ui
docker push ${UI_IMG}
docker push ${DB_MANAGER_IMG}
docker push ${CONTROLLER_IMG}

# algorithm server
docker push ${ALGORITHM_IMG}

# http client
docker push ${CLIENT_IMG}

echo -e "\n Docker images push succeeded\n"
