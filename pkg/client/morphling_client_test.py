from __future__ import print_function

import api_pb2
import api_pb2_grpc
import grpc


def main():

    manager_server = "morphling-db-manager:6799"
    channel_manager = grpc.insecure_channel(manager_server)
    qps_previous = int(1)
    print(qps_previous)

    mls = []
    ml = api_pb2.KeyValue(key="qps", value=str(qps_previous))
    mls.append(ml)

    stub_ = api_pb2_grpc.DBStub(channel_manager)
    result = stub_.SaveResult(
        api_pb2.SaveResultRequest(
            trial_name="test-trial", namespace="test-namespace", results=mls
        ),
        timeout=20,
    )


if __name__ == "__main__":
    main()
