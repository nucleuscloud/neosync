# Neosync TypeScript SDK

This SDK contains the generated types for Neosync API.
This SDK is dogfooded by the main Neosync webapp to ensure its durability.

## Installation

```sh
npm install @neosync/sdk @bufbuild/protobuf
```

## Detailed Docs

This README shows the basics of how to use the SDK.

For more detailed docs, go [here](https://docs.neosync.dev/api/typescript).

## Usage

For a prime example of how to us this SDK, view the [withNeosyncContext](https://github.com/nucleuscloud/neosync/blob/main/frontend/apps/web/api-only/neosync-context.ts#L23) method in the Neosync app's BFF layer.

### Note on Transports

Based on your usage, you'll have to install a different version of `connect` to provide the correct Transport based on your environment.

- Node: [@connectrpc/connect-node](https://connectrpc.com/docs/node/using-clients)
- Web: [@connectrpc/connect-web](https://connectrpc.com/docs/web/using-clients)

Install whichever one makes sense for you

```sh
npm install @connectrpc/connect-node
npm install @connectrpc/connect-web
```

Neosync API serves up `Connect`, which can listen using Connect, gRPC, or Web protocols.
Each of the libraries above provides all three of those protocols, but it's recommended to use `createConnectTransport` for the most efficient setup.

```ts
import { getNeosyncClient } from '@neosync/sdk';
import { createConnectTransport } from '@connectrpc/connect-node';

const neosyncClient = getNeosyncClient({
  getTransport(interceptors) {
    return createConnectTransport({
      baseUrl: '<url>',
      httpVersion: '2',
      interceptors: interceptors,
    });
  },
});
```

## Authenticating

To authenticate the TS Neosync Client, a function may be provided to the configuration that will be invoked prior to every request.
This gives flexability in how the access token may be retrieved and supports either a Neosync API Key or a standard user JWT token.

When the `getAccessToken` function is provided, the Neosync Client is configured with an auth interceptor that attaches the `Authorization` header to every outgoingn request with the access token returned from the function.
This is why the `getTransport` method receives a list of interceptors, and why it's important to hook them up to pass them through to the relevant transport being used.

```ts
import { getNeosyncClient } from '@neosync/sdk';
import { createConnectTransport } from '@connectrpc/connect-node';

const neosyncClient = getNeosyncClient({
  getAccessToken: () => process.env.NEOSYNC_API_KEY,
  getTransport(interceptors) {
    return createConnectTransport({
      baseUrl: process.env.NEOSYNC_API_URL,
      httpVersion: '2',
      interceptors: interceptors,
    });
  },
});
```

### Neosync App

In the Neosync dashboard app, we pull the user access token off of the incoming request (auth is configured using `next-auth`.).
This way we can ensure that all requests are using the user's access token and are passed through to Neosync API.
