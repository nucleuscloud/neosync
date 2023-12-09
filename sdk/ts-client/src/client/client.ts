import {
  Interceptor,
  PromiseClient,
  Transport,
  createPromiseClient,
} from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-node";

import { ApiKeyService } from "./mgmt/v1alpha1/api_key_connect";
import { ConnectionService } from "./mgmt/v1alpha1/connection_connect";
import { JobService } from "./mgmt/v1alpha1/job_connect";
import { TransformersService } from "./mgmt/v1alpha1/transformer_connect";
import { UserAccountService } from "./mgmt/v1alpha1/user_account_connect";

export type NeosyncClient = NeosyncV1alpha1Client;
export type ClientVersion = "v1alpha1" | "latest";

export interface NeosyncV1alpha1Client {
  connections: PromiseClient<typeof ConnectionService>;
  users: PromiseClient<typeof UserAccountService>;
  jobs: PromiseClient<typeof JobService>;
  transformers: PromiseClient<typeof TransformersService>;
  apikeys: PromiseClient<typeof ApiKeyService>;
}

/**
 * Function that returns the access token either as a string or a string promise
 */
export type GetAccessTokenFn = () => string | Promise<string>;

export interface ClientConfig {
  /**
   * The baseurl for Neosync API
   */
  apiBaseUrl: string;

  /**
   * Return the access token to be used for authenticating against Neosync API
   * This will either be a JWT, or an API Key
   * It will be used to construct the Authorization Header in the format: Authorization: Bearer <access token>
   */
  getAccessToken?: GetAccessTokenFn;
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
  version: "latest"
): NeosyncClient;
/**
 * Returns the v1alpha1 version of the Neosync Client
 */
export function getNeosyncClient(
  config: ClientConfig,
  version: "v1alpha1"
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
  const transport = getConnectTransport(
    config.apiBaseUrl,
    config.getAccessToken
  );
  return {
    connections: createPromiseClient(ConnectionService, transport),
    users: createPromiseClient(UserAccountService, transport),
    jobs: createPromiseClient(JobService, transport),
    transformers: createPromiseClient(TransformersService, transport),
    apikeys: createPromiseClient(ApiKeyService, transport),
  };
}

function getConnectTransport(
  baseUrl: string,
  getAccessToken?: GetAccessTokenFn
): Transport {
  const interceptors = getAccessToken
    ? [getAuthInterceptor(getAccessToken)]
    : [];
  return createConnectTransport({
    baseUrl,
    httpVersion: "2",
    interceptors: interceptors,
  });
}

function getAuthInterceptor(getAccessToken: GetAccessTokenFn): Interceptor {
  return (next) => async (req) => {
    const accessToken = await getAccessToken();
    req.header.set("Authorization", `Bearer ${accessToken}`);
    return next(req);
  };
}
