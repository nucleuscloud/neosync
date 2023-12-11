# Neosync TypeScript SDK

This SDK contains the generated types for Neosync API.
This SDK is dogfooded by the main Neosync webapp to ensure its durability.

## Installation

```sh
npm install @neosync/sdk
```

## Usage

For a prime example of how to us this SDK, view the [withNeosyncContext](https://github.com/nucleuscloud/neosync/blob/main/frontend/apps/web/api-only/neosync-context.ts#L23) method in the Neosync app's BFF layer.


### Note on Transports
Based on your usage, you'll have to install a different version of `connect` to provide the correct Transport based on your environment.

* Node: [@connectrpc/connect-node](https://connectrpc.com/docs/node/using-clients)
* Web: [@connectrpc/connect-web](https://connectrpc.com/docs/web/using-clients)

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
    })
  }
});
```
