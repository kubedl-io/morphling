#!/bin/bash
export PROTO_FILE=api.proto
protoc --go_out=plugins=grpc:./ "$PROTO_FILE"
python3 -m grpc_tools.protoc -I. --python_out=python3 --grpc_python_out=python3 "$PROTO_FILE"