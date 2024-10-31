#!/bin/bash

# Set the proto file name
export PROTO_FILE=predict.proto

# Generate Go code
protoc --go_out=. "$PROTO_FILE"

# Generate Python code
python3 -m grpc_tools.protoc -I. --python_out=python3 --grpc_python_out=python3 "$PROTO_FILE" 

# Output completion information
echo "gRPC code generation completed for $PROTO_FILE"