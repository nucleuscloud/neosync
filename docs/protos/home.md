---
title: APIs and SDKs
description: Learn about Neosync's APIs and SDKs in order to anonymize data and generate synthetic data
hide_title: false
id: home
slug: /
---

## Introduction

Neosync has first-class APIs and SDKs that developers can use to integrate Neosync into their workflow. To learn more about how the Neosync API fits into the overall architecture, check out the Check out the [platform page](/platform).

Neosync API serves up [Connect](https://github.com/connectrpc), which can listen using Connect, gRPC, or HTTP protocols. All of our APIs are generated from Protobuf files and our SDKs call Connect endpoints by default. Each SDK can be configured to use gRPC or REST in lieu of the default (Connect).

## Configuration

There are a few inputs that any SDK needs in order to be properly configured.

1. API URL
2. Account ID
3. API Key (required for Neosync Cloud or self-hosted authenticated environments)

### API Url

If using Neosync Cloud, the backend api url is: `https://neosync-api.svcs.neosync.dev`

The standard localhost url is: `http://localhost:8080`

### Account ID

The account ID is necessary for some requests that do not have an obvious identifier like retrieving a list of jobs, or a list of connections.
This can be found by going into the app on the `/:accountName/settings` page and found in the header.

### API Key

An access token (api key, or user jwt) must be used to access authenticated Neosync environments.
For an API Key, this can be created at `/:accountName/settings/api-keys`.

## Clients

### Go

The Go SDK that is committed to the Neosync repo may be freely imported and utilized.
See the [Go SDK](./go.md) page for more information on how to use the SDK.

All of the generated code lives [here](https://github.com/nucleuscloud/neosync/tree/main/backend/gen/go/protos/mgmt/v1alpha1).

### TypeScript

The TypeScript SDK is published to the `npm` registry. It is generated from Neosync protos and is used by the Neosync App.
See the [TS SDK](./typescript.md) page for more information on how to use the SDK.

All of the generated code lives [here](https://github.com/nucleuscloud/neosync/tree/main/frontend/packages/sdk).

### Protos

All of Neosync's protos are public and can be found [here](https://github.com/nucleuscloud/neosync/tree/main/backend/protos).
A new SDK can be easily generated by augmenting the `buf.gen.yaml` file, or providing a separate one when running the `buf` cli to generate a different SDK for other purposes.