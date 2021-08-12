import time
from concurrent import futures

import grpc

from api.v1alpha1.grpc_proto.grpc_algorithm.python3 import api_pb2_grpc
from api.v1alpha1.grpc_proto.health.python import health_pb2_grpc
from pkg.algorithm.v1alpha1.grid.service import BaseService

_ONE_DAY_IN_SECONDS = 60 * 60 * 24
DEFAULT_PORT = "0.0.0.0:9996"


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service = BaseService()
    api_pb2_grpc.add_SuggestionServicer_to_server(service, server)
    health_pb2_grpc.add_HealthServicer_to_server(service, server)
    server.add_insecure_port(DEFAULT_PORT)
    print("Listening...")
    server.start()
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == "__main__":
    serve()
