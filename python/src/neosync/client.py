from grpc import (
    Channel,
    secure_channel,
    insecure_channel,
    ssl_channel_credentials,
    intercept_channel,
    UnaryUnaryClientInterceptor,
    UnaryStreamClientInterceptor,
    StreamUnaryClientInterceptor,
    StreamStreamClientInterceptor,
)
from typing import Callable, Optional, Union
from neosync.mgmt.v1alpha1 import (
    anonymization_pb2_grpc,
    api_key_pb2_grpc,
    connection_data_pb2_grpc,
    connection_pb2_grpc,
    job_pb2_grpc,
    metrics_pb2_grpc,
    transformer_pb2_grpc,
    user_account_pb2_grpc,
)

# Function that returns the access token
GetAccessTokenFn = Callable[[], Union[str, None]]


class Neosync:
    """A client for interacting with the Neosync API.

    This class provides access to various Neosync services including connections,
    jobs, metrics, transformers, and more. It handles authentication and
    communication with the Neosync API endpoints.

    Args:
        access_token (Optional[str]): A static bearer token for API authentication.
            Mutually exclusive with get_access_token.
        api_url (Optional[str]): The URL of the Neosync API endpoint.
            Defaults to "neosync-api.svcs.neosync.dev:443".
        get_access_token (Optional[GetAccessTokenFn]): A callback function that returns
            a bearer token for API authentication. Mutually exclusive with access_token.
        insecure (Optional[bool]): If True, creates an insecure channel without TLS.
            Defaults to False.

    Attributes:
        connectiondata: Service client for connection data operations.
        connections: Service client for managing connections.
        jobs: Service client for managing jobs.
        metrics: Service client for accessing metrics.
        transformers: Service client for transformer operations.
        users: Service client for user account management.
        anonymization: Service client for anonymization operations.
        apikeys: Service client for API key management.

    Example:
        >>> client = Neosync(access_token="your-token")
        >>> # Access various services
        >>> jobs = client.jobs.ListJobs(GetJobsRequest(account_id="your-account-id"))
    """

    def __init__(
        self,
        access_token: Optional[str] = None,
        api_url: Optional[str] = "neosync-api.svcs.neosync.dev:443",
        get_access_token: Optional[GetAccessTokenFn] = None,
        insecure: Optional[bool] = False,
    ):
        config = _ClientConfig(access_token, api_url, get_access_token, insecure)
        channel = _get_channel_from_config(config)
        self.connectiondata = connection_data_pb2_grpc.ConnectionDataServiceStub(
            channel
        )
        self.connections = connection_pb2_grpc.ConnectionServiceStub(channel)
        self.jobs = job_pb2_grpc.JobServiceStub(channel)
        self.metrics = metrics_pb2_grpc.MetricsServiceStub(channel)
        self.transformers = transformer_pb2_grpc.TransformersServiceStub(channel)
        self.users = user_account_pb2_grpc.UserAccountServiceStub(channel)
        self.anonymization = anonymization_pb2_grpc.AnonymizationServiceStub(channel)
        self.apikeys = api_key_pb2_grpc.ApiKeyServiceStub(channel)


class _ClientConfig:
    def __init__(
        self,
        access_token: Optional[str] = None,
        api_url: Optional[str] = None,
        get_access_token: Optional[GetAccessTokenFn] = None,
        insecure: Optional[bool] = False,
    ):
        if access_token is not None:
            self.get_access_token = lambda: access_token
        elif get_access_token is not None:
            self.get_access_token = get_access_token
        else:
            self.get_access_token = None

        self.api_url = api_url
        self.insecure = insecure


# Pulled from grpc examples: https://github.com/grpc/grpc/blob/master/examples/python/interceptors/headers/generic_client_interceptor.py
class _GenericClientInterceptor(
    UnaryUnaryClientInterceptor,
    UnaryStreamClientInterceptor,
    StreamUnaryClientInterceptor,
    StreamStreamClientInterceptor,
):
    def __init__(self, interceptor_function):
        self._fn = interceptor_function

    def intercept_unary_unary(self, continuation, client_call_details, request):
        new_details, new_request_iterator, postprocess = self._fn(
            client_call_details, iter((request,)), False, False
        )
        response = continuation(new_details, next(new_request_iterator))
        return postprocess(response) if postprocess else response

    def intercept_unary_stream(self, continuation, client_call_details, request):
        new_details, new_request_iterator, postprocess = self._fn(
            client_call_details, iter((request,)), False, True
        )
        response_it = continuation(new_details, next(new_request_iterator))
        return postprocess(response_it) if postprocess else response_it

    def intercept_stream_unary(
        self, continuation, client_call_details, request_iterator
    ):
        new_details, new_request_iterator, postprocess = self._fn(
            client_call_details, request_iterator, True, False
        )
        response = continuation(new_details, new_request_iterator)
        return postprocess(response) if postprocess else response

    def intercept_stream_stream(
        self, continuation, client_call_details, request_iterator
    ):
        new_details, new_request_iterator, postprocess = self._fn(
            client_call_details, request_iterator, True, True
        )
        response_it = continuation(new_details, new_request_iterator)
        return postprocess(response_it) if postprocess else response_it


def _get_auth_interceptor(
    get_access_token: Optional[GetAccessTokenFn] = None,
) -> _GenericClientInterceptor:
    def interceptor(
        client_call_details,
        request_iterator,
        request_streaming,
        response_streaming,
    ):
        metadata = []
        if client_call_details.metadata is not None:
            metadata = list(client_call_details.metadata)

        if get_access_token:
            token = get_access_token()
            if token:
                metadata.append(("authorization", f"Bearer {token}"))

        client_call_details = client_call_details._replace(metadata=metadata)
        return client_call_details, request_iterator, None

    return _GenericClientInterceptor(interceptor)


def _get_channel_from_config(config: _ClientConfig) -> Channel:
    """Returns a gRPC channel from the client configuration"""
    interceptors = (
        [_get_auth_interceptor(config.get_access_token)]
        if config.get_access_token
        else []
    )
    if config.insecure:
        return intercept_channel(insecure_channel(config.api_url), *interceptors)
    else:
        return intercept_channel(
            secure_channel(config.api_url, ssl_channel_credentials()), *interceptors
        )
