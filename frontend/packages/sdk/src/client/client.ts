import {
  Interceptor,
  PromiseClient,
  Transport,
  createPromiseClient,
} from '@connectrpc/connect';

import { AnonymizationService } from './mgmt/v1alpha1/anonymization_connect.js';
import { ApiKeyService } from './mgmt/v1alpha1/api_key_connect.js';
import { ConnectionService } from './mgmt/v1alpha1/connection_connect.js';
import { ConnectionDataService } from './mgmt/v1alpha1/connection_data_connect.js';
import { JobService } from './mgmt/v1alpha1/job_connect.js';
import { MetricsService } from './mgmt/v1alpha1/metrics_connect.js';
import { TransformersService } from './mgmt/v1alpha1/transformer_connect.js';
import { UserAccountService } from './mgmt/v1alpha1/user_account_connect.js';

export type NeosyncClient = NeosyncV1alpha1Client;
export type ClientVersion = 'v1alpha1' | 'latest';

export interface NeosyncV1alpha1Client {
  connections: PromiseClient<typeof ConnectionService>;
  users: PromiseClient<typeof UserAccountService>;
  jobs: PromiseClient<typeof JobService>;
  transformers: PromiseClient<typeof TransformersService>;
  apikeys: PromiseClient<typeof ApiKeyService>;
  connectiondata: PromiseClient<typeof ConnectionDataService>;
  metrics: PromiseClient<typeof MetricsService>;
  anonymization: PromiseClient<typeof AnonymizationService>;
}

/**
 * Function that returns the access token either as a string or a string promise
 */
export type GetAccessTokenFn = () => string | Promise<string>;

export interface ClientConfig {
  /**
   * Return the access token to be used for authenticating against Neosync API
   * This will either be a JWT, or an API Key
   * It will be used to construct the Authorization Header in the format: Authorization: Bearer <access token>
   */
  getAccessToken?: GetAccessTokenFn;
  /**
   * Return the connect transport for the appropriate environment (connect, grpc, web)
   * @param interceptors - A list of interceptors that have been pre-compuled. If `getAccessToken` is provided, this will include the auth interceptor
   */
  getTransport(interceptors: Interceptor[]): Transport;
}

/**
 * Returns the latest version of the Neosync Client
 */
export function getNeosyncClient(config: ClientConfig): NeosyncClient;
/**
 * Returns the latest version of the Neosync Client
 */
export function getNeosyncClient(
  config: ClientConfig,
  version: 'latest'
): NeosyncClient;
/**
 * Returns the v1alpha1 version of the Neosync Client
 */
export function getNeosyncClient(
  config: ClientConfig,
  version: 'v1alpha1'
): NeosyncV1alpha1Client;
export function getNeosyncClient(
  config: ClientConfig,
  _version?: ClientVersion
): NeosyncClient {
  return getNeosyncV1alpha1Client(config);
}

/**
 * Returns the v1alpha1 version of the Neosync client
 * @returns
 */
export function getNeosyncV1alpha1Client(
  config: ClientConfig
): NeosyncV1alpha1Client {
  const interceptors = config.getAccessToken
    ? [getAuthInterceptor(config.getAccessToken)]
    : [];
  const transport = config.getTransport(interceptors);
  return {
    connections: createPromiseClient(ConnectionService, transport),
    users: createPromiseClient(UserAccountService, transport),
    jobs: createPromiseClient(JobService, transport),
    transformers: createPromiseClient(TransformersService, transport),
    apikeys: createPromiseClient(ApiKeyService, transport),
    connectiondata: createPromiseClient(ConnectionDataService, transport),
    metrics: createPromiseClient(MetricsService, transport),
    anonymization: createPromiseClient(AnonymizationService, transport),
  };
}

function getAuthInterceptor(getAccessToken: GetAccessTokenFn): Interceptor {
  return (next) => async (req) => {
    const accessToken = await getAccessToken();
    req.header.set('Authorization', `Bearer ${accessToken}`);
    return next(req);
  };
}
