---
title: Neosync for Go
description: Learn about Neosync's Go SDK and how you can use it to anonymize data and generate synthetic data
id: go
hide_title: false
slug: /go
---

## Introduction

The Neosync Go SDK is publicly available and can be added to any Go-based project by running `go get github.com/nucleuscloud/neosync`.

Neosync's CLI is the primary user of the Go SDK today, and can be referenced for examples of how to use the SDK.

## Configuration

There are a few inputs that the SDK needs in order to be properly configured.

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

## Getting Started

Neosync is made up of a number of different services that live inside of the same process.
They are roughly split up in terms of their resource types, and correspond nicely with what resources that are found in the web application.

We can initialize the job service client to trigger a new job run like so:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

func main() {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		os.Getenv("NEOSYNC_API_URL"),
	)

	_, err := jobclient.CreateJobRun(context.Background(), connect.NewRequest(&mgmtv1alpha1.CreateJobRunRequest{
		JobId: "<neosync-job-id>",
	}))
	if err != nil {
		panic(err)
	}
}
```

## Go SDK Packages

There are two packages that are made available for connecting to a Neosync API.

- `github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1`
- `github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect`

The first package is made up of the generated types. This includes all of the request, response, and DTO types.

The second package are where the client and server structs live for use with either creating a Neosync Client, or Neosync Server.

## Adding Authentication to the SDK

If developing Neosync locally, it's easy to run in the default mode that doesn't require authentication.
However, if interacting with a production environment, it most likely will require authentication like an API Key.

Providing the authentication header can be a little wonky in Go and can be done in two different ways.

1. Attaching to the HTTP client
2. Providing an interceptor to the SDK Clients that patch in the header on every request.

The example below shows how to augment the HTTP Client to include the header:

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

func main() {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		newHttpClient(map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", os.Getenv("NEOSYNC_API_KEY")),
		}),
		os.Getenv("NEOSYNC_API_URL"),
	)
	_ = jobclient
}

func newHttpClient(
	headers map[string]string,
) *http.Client {
	return &http.Client{
		Transport: &headerTransport{
			Transport: http.DefaultTransport,
			Headers:   headers,
		},
	}
}

type headerTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header == nil {
		req.Header = http.Header{}
	}
	for key, value := range t.Headers {
		req.Header.Add(key, value)
	}
	return t.Transport.RoundTrip(req)
}
```
