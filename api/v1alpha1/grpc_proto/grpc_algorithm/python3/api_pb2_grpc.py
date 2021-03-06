# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import api_pb2 as api__pb2
import grpc


class SuggestionStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GetSuggestions = channel.unary_unary(
            "/api.suggestion.Suggestion/GetSuggestions",
            request_serializer=api__pb2.SamplingRequest.SerializeToString,
            response_deserializer=api__pb2.SamplingResponse.FromString,
        )
        self.ValidateAlgorithmSettings = channel.unary_unary(
            "/api.suggestion.Suggestion/ValidateAlgorithmSettings",
            request_serializer=api__pb2.SamplingValidationRequest.SerializeToString,
            response_deserializer=api__pb2.SamplingValidationResponse.FromString,
        )


class SuggestionServicer(object):
    """Missing associated documentation comment in .proto file."""

    def GetSuggestions(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")

    def ValidateAlgorithmSettings(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Method not implemented!")
        raise NotImplementedError("Method not implemented!")


def add_SuggestionServicer_to_server(servicer, server):
    rpc_method_handlers = {
        "GetSuggestions": grpc.unary_unary_rpc_method_handler(
            servicer.GetSuggestions,
            request_deserializer=api__pb2.SamplingRequest.FromString,
            response_serializer=api__pb2.SamplingResponse.SerializeToString,
        ),
        "ValidateAlgorithmSettings": grpc.unary_unary_rpc_method_handler(
            servicer.ValidateAlgorithmSettings,
            request_deserializer=api__pb2.SamplingValidationRequest.FromString,
            response_serializer=api__pb2.SamplingValidationResponse.SerializeToString,
        ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
        "api.suggestion.Suggestion", rpc_method_handlers
    )
    server.add_generic_rpc_handlers((generic_handler,))


# This class is part of an EXPERIMENTAL API.
class Suggestion(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def GetSuggestions(
        request,
        target,
        options=(),
        channel_credentials=None,
        call_credentials=None,
        insecure=False,
        compression=None,
        wait_for_ready=None,
        timeout=None,
        metadata=None,
    ):
        return grpc.experimental.unary_unary(
            request,
            target,
            "/api.suggestion.Suggestion/GetSuggestions",
            api__pb2.SamplingRequest.SerializeToString,
            api__pb2.SamplingResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
        )

    @staticmethod
    def ValidateAlgorithmSettings(
        request,
        target,
        options=(),
        channel_credentials=None,
        call_credentials=None,
        insecure=False,
        compression=None,
        wait_for_ready=None,
        timeout=None,
        metadata=None,
    ):
        return grpc.experimental.unary_unary(
            request,
            target,
            "/api.suggestion.Suggestion/ValidateAlgorithmSettings",
            api__pb2.SamplingValidationRequest.SerializeToString,
            api__pb2.SamplingValidationResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
        )
