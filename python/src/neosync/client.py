from grpc import Channel, Interceptor, secure_channel, insecure_channel, ssl_channel_credentials
from dataclasses import dataclass
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

#  Function that returns the access token
GetAccessTokenFn = Callable[[], Union[str, None]]

@dataclass
class NeosyncV1alpha1Client:
  connectiondata: connection_data_pb2_grpc.ConnectionDataServiceStub
  connections: connection_pb2_grpc.ConnectionServiceStub
  jobs: job_pb2_grpc.JobServiceStub
  metrics: metrics_pb2_grpc.MetricsServiceStub
  transformers: transformer_pb2_grpc.TransformersServiceStub
  users: user_account_pb2_grpc.UserAccountServiceStub
  anonymization: anonymization_pb2_grpc.AnonymizationServiceStub
  apikeys: api_key_pb2_grpc.ApiKeyServiceStub

class ClientConfig:
  def __init__(
      self,
      access_token: Optional[str] = None,
      api_url: Optional[str] = "neosync-api.svcs.neosync.dev:443",
      get_access_token: Optional[GetAccessTokenFn] = None,
      insecure: Optional[bool] = False,
  ):
    if access_token is not None:
      self.get_access_token = lambda: access_token
    elif get_access_token is not None:
      self.get_access_token = get_access_token

    self.api_url = api_url
    self.insecure = insecure

def get_neosync_client(
    config: ClientConfig,
    version: Optional[str] = "latest",
) -> NeosyncV1alpha1Client:
  """Returns a Neosync client instance.

  Args:
    config: The client configuration
    version: The version of the client to return. Defaults to "latest"
  """
  if version not in ['latest', 'v1alpha1']:
      raise ValueError("Version must be either 'latest' or 'v1alpha1'")

  return _get_neosync_v1alpha1_client(config)

def _get_neosync_v1alpha1_client(config: ClientConfig) -> NeosyncV1alpha1Client:
  """Returns the v1alpha1 version of the Neosync client"""
  channel = _get_channel_from_config(config)
  return NeosyncV1alpha1Client(
    connectiondata=connection_data_pb2_grpc.ConnectionDataServiceStub(channel),
    connections=connection_pb2_grpc.ConnectionServiceStub(channel),
    jobs=job_pb2_grpc.JobServiceStub(channel),
    metrics=metrics_pb2_grpc.MetricsServiceStub(channel),
    transformers=transformer_pb2_grpc.TransformersServiceStub(channel),
    users=user_account_pb2_grpc.UserAccountServiceStub(channel),
    anonymization=anonymization_pb2_grpc.AnonymizationServiceStub(channel),
    apikeys=api_key_pb2_grpc.ApiKeyServiceStub(channel),
  )

def _get_auth_interceptor(get_access_token: Optional[GetAccessTokenFn] = None) -> Interceptor:
    def interceptor(continuation, client_call_details, request):
        if get_access_token:
            token = get_access_token()
            if token:
                metadata = []
                if client_call_details.metadata is not None:
                    metadata = list(client_call_details.metadata)
                metadata.append(('authorization', f'Bearer {token}'))
                client_call_details = client_call_details._replace(metadata=metadata)

        return continuation(client_call_details, request)

    return interceptor

def _get_channel_from_config(config: ClientConfig) -> Channel:
  """Returns a gRPC channel from the client configuration"""
  interceptors = [_get_auth_interceptor(config.get_access_token)] if config.get_access_token else []
  if config.insecure:
    return insecure_channel(config.api_url).intercept_channel(interceptors)
  else:
    return secure_channel(config.api_url, ssl_channel_credentials()).intercept_channel(interceptors)
