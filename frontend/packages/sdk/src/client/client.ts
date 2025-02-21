import {
  Client,
  Interceptor,
  Transport,
  createClient,
} from '@connectrpc/connect';

import { AccountHookService } from './mgmt/v1alpha1/account_hook_pb.js';
import { AnonymizationService } from './mgmt/v1alpha1/anonymization_pb.js';
import { ApiKeyService } from './mgmt/v1alpha1/api_key_pb.js';
import { ConnectionDataService } from './mgmt/v1alpha1/connection_data_pb.js';
import { ConnectionService } from './mgmt/v1alpha1/connection_pb.js';
import { JobService } from './mgmt/v1alpha1/job_pb.js';
import { MetricsService } from './mgmt/v1alpha1/metrics_pb.js';
import { TransformersService } from './mgmt/v1alpha1/transformer_pb.js';
import { UserAccountService } from './mgmt/v1alpha1/user_account_pb.js';
export type NeosyncClient = NeosyncV1alpha1Client;
export type ClientVersion = 'v1alpha1' | 'latest';

export interface NeosyncV1alpha1Client {
  connections: Client<typeof ConnectionService>;
  users: Client<typeof UserAccountService>;
  jobs: Client<typeof JobService>;
  transformers: Client<typeof TransformersService>;
  apikeys: Client<typeof ApiKeyService>;
  connectiondata: Client<typeof ConnectionDataService>;
  metrics: Client<typeof MetricsService>;
  anonymization: Client<typeof AnonymizationService>;
  accountHooks: Client<typeof AccountHookService>;
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
    connections: createClient(ConnectionService, transport),
    users: createClient(UserAccountService, transport),
    jobs: createClient(JobService, transport),
    transformers: createClient(TransformersService, transport),
    apikeys: createClient(ApiKeyService, transport),
    connectiondata: createClient(ConnectionDataService, transport),
    metrics: createClient(MetricsService, transport),
    anonymization: createClient(AnonymizationService, transport),
    accountHooks: createClient(AccountHookService, transport),
  };
}

function getAuthInterceptor(getAccessToken: GetAccessTokenFn): Interceptor {
  return (next) => async (req) => {
    const accessToken = await getAccessToken();
    req.header.set('Authorization', `Bearer ${accessToken}`);
    return next(req);
  };
}
