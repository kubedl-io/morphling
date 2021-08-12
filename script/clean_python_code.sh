#!/usr/bin/env bash

#set -e

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
cd ${SCRIPT_ROOT}
echo "cd to ${SCRIPT_ROOT}"

#black
for file in $(ls -d */);
do
black ${file}
pycln ${file}
isort ${file}
echo ${file%%/};
done
