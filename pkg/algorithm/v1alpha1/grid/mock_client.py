# encoding: utf-8
import logging
import os

import grpc

from api.v1alpha1.grpc_proto.grpc_algorithm.python3 import api_pb2_grpc, api_pb2

logger = logging.getLogger('grpc_algorithm-client')
logger.setLevel(logging.INFO)
console = logging.StreamHandler()
console.setLevel(logging.INFO)
formatter = logging.Formatter('%(asctime)s [%(levelname)s] - %(message)s')
console.setFormatter(formatter)
logger.addHandler(console)


def validate(stub):

    # parameter_type = api_pb2.ParameterType(CATEGORICAL)

    par_1 = api_pb2.ParameterSpec(name="cpu", parameter_type="CATEGORICAL", feasible_space=["1", "2", "3.5"])
    par_2 = api_pb2.ParameterSpec(name="memory", parameter_type="CATEGORICAL", feasible_space=["10", "20", "35"])

    parameters = [par_1, par_2]
    request = api_pb2.SamplingValidationRequest(algorithm_name="grid", sampling_number_specified=3, is_maximize=True, parameters=parameters)
    logger.info("validate", request)
    response = stub.ValidateAlgorithmSettings(request=request)


def grpc_server():
    server = os.getenv("GRPC_SERVER")
    if server:
        return server
    else:
        return "localhost"


# def random_id(end):
#     return str(random.randint(0, end))
#
#
# def generate_request(method_name):
#     for _ in range(0, 3):
#         request = api_pb2.TalkRequest(data=random_id(5), meta="PYTHON")
#         logger.info("%s data:%s,meta:%s", method_name, request.data, request.meta)
#         yield request
#         time.sleep(random.uniform(0.5, 1.5))


def print_response(method_name, response):
    for result in response.results:
        kv = result.kv
        logger.info("%s [%d] %d [%s %s %s,%s:%s]", method_name,
                    response.status, result.id, kv["meta"], result.type, kv["id"], kv["idx"], kv["data"])


def run():
    address = grpc_server() + ":9996"
    channel = grpc.insecure_channel(address)
    stub = api_pb2_grpc.SuggestionStub(channel)
    logger.info("Unary RPC")
    validate(stub)


if __name__ == '__main__':
    run()
