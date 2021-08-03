import logging

import grpc

from api.v1alpha1.grpc_proto.grpc_algorithm.python3 import api_pb2_grpc, api_pb2
from pkg.algorithm.v1alpha1.grid.base_service import BaseSamplingService
from pkg.algorithm.v1alpha1.internal.base_health_service import HealthServicer
logger = logging.getLogger(__name__)

support_algorithms = ["grid", "random"]


def _set_validate_context_error(context, error_message):
    context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
    context.set_details(error_message)
    logger.info(error_message)
    print(error_message)
    return api_pb2.SamplingValidationResponse()


class BaseService(api_pb2_grpc.SuggestionServicer, HealthServicer):
    def __init__(self):
        super(BaseService, self).__init__()

    def ValidateAlgorithmSettings(self, request, context):
        algorithm_name = request.algorithm_name
        if algorithm_name not in support_algorithms:
            return _set_validate_context_error(context, "algorithm {} is not supported".format(algorithm_name))
        return api_pb2.SamplingValidationResponse()

    def GetSuggestions(self, request, context):

        if request.algorithm_name in support_algorithms:
            service = BaseSamplingService(request)
            if request.required_sampling + int(len(request.existing_results)) > min(service.space_size, request.sampling_number_specified):
                return _set_validate_context_error(context, "space size {} is not enough to provide another {} samplings".format(min(service.space_size, request.sampling_number_specified), request.required_sampling))
            new_assignments = service.get_assignment(request)
            return api_pb2.SamplingResponse(assignments_set=new_assignments)
        else:
            return _set_validate_context_error(context, "algorithm {} is not supported".format(request.algorithm_name))
