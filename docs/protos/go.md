---
title: Neosync for Go
description: Learn about Neosync's Go SDK and how you can use it to anonymize data and generate synthetic data
id: go
hide_title: false
slug: /go
---

## Introduction

The Neosync Go SDK is publicly available and can be added to any Go project. With the Neosync Go SDK, you can:

1. Anonymize structured data and generate synthetic data
2. Anonymize free-form text data
3. Create resources in Neosync such as Jobs, Connections, Transformers and more

## Installation

You can add the Neosync Go SDK using:

`go get github.com/nucleuscloud/neosync`.

## Prerequisites

There are a few prerequisites that the SDK needs in order to be properly configured.

| **Properties** | **Details**                                                                                                                                                                 |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **API URL**    | Production: `https://neosync-api.svcs.neosync.dev`<br /> Local: `http://localhost:8080`                                                                                     |
| **Account ID** | The account ID may be necessary for some requests and can be found by going into the `/:accountName/settings` page in the Neosync App                                       |
| **API Key**    | An access token (API key, or user JWT) must be used to access authenticated Neosync environments. For an API Key, this can be created at `/:accountName/settings/api-keys`. |

## Authentication

If you are using Neosync locally and in unauthenticated mode then there is no authentication required and you can move onto the [Getting Started](go#getting-started) section.

If you are using Neosync locally in `auth mode` or using Neosync Cloud, you can authenticate with the Neosync server using an API URL and API Key. There are two ways to provide the authentication header.

1. Attaching to the HTTP client
2. Providing an interceptor to the SDK Clients that patch in the header on every request.

The example below shows the first option which attaches the API Key to the HTTP Client header:

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

func main() {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		newHttpClient(map[string]string{ // create an instance of a newHttpClient
			"Authorization": fmt.Sprintf("Bearer %s", os.GetEnv("API_KEY")), // pass in the API_KEY through an environment variable
		}),
		os.GetEnv("API_URL"), // pass in the API_URL through an environment variable
	)

	// rest of code to call an API in the JobServiceClient goes here
	// ...
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

## Getting started

In this section, we're going to walk through two examples that show you how to make an API call using Neosync's GO SDK. For a complete list of the APIs, check out the APIs in the `Services` section of our [protos](/api/mgmt/v1alpha1/job.proto#jobservice).

Neosync is made up of a number of different services that live inside of the same process. In order to connect to the Neosync API and use the services, we make two packages available:

- `github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1`
- `github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect`

The first package is made up of the generated types. This includes all of the request, response, and DTO types.

The second package are where the client and server structs live for use with either creating a Neosync Client, or Neosync Server.

### Anonymizing Structured Data

A straightforward use case is to anonymize sensitive data in an API request. Let's look at an example.

First, let's

```json
// input
{
  "user": {
    "name": "John Doe",
    "email": "john@example.com"
  },
  "details": {
    "address": "123 Main St",
    "phone": "555-1234",
    "favorites": ["dog", "cat", "bird"],
    "name": "jake"
  }
}
```

Our input object is a simple user's object that we may get through a user sign up flow. In this object, we have a few sensitive fields that we want to anonymize: `name`, `email`, `address` and `phone`. We can leave the `favorites` as-is for now.

In order to anonymize this object, you can use Neosync's `AnonymizeSingle` API to send in a single object with sensitive data and get back an anonymized version of that object. You have full control over how you anonymize the data or generate new synthetic data. Here's how you do it:

```go
// transformer definitions
transformerMappings := []*mgmtv1alpha1.TransformerMapping{
  {
    Expression: `(.. | objects | select(has("name")) | .name)`, // find and transform all name fields in objects
    Transformer: &mgmtv1alpha1.TransformerConfig{
      Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
        TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
          PreserveLength: true,
        },
      },
    },
  },
  {
    Expression: `.user.email`, // transform user.email field
    Transformer: &mgmtv1alpha1.TransformerConfig{
      Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
        TransformEmailConfig: &mgmtv1alpha1.TransformEmail{},
      },
    },
  },
   {
    Expression: `.details.favorites[]`, // transform each element in details.favorite array
    Transformer: &mgmtv1alpha1.TransformerConfig{
      Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
        TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
      },
    },
  },
}
```

Let's take a closer look at what we're doing here. Neosync's AnonymizeSingle API uses [JQ](https://jqlang.github.io/jq/manual/) expressions to target field(s) in your object. This means that you don't have to parse your object before sending it to Neosync. You can pass it in as-is and just write JQ expressions to target the field(s) that you want to anonymize or generate.

Our output will look something like this:

```json
// output
{
  "user": {
    "name": "Bob Smith",
    "email": "random@different.com"
  },
  "details": {
    "address": "123 Main St",
    "phone": "555-1234",
    "favorites": ["sfgjr", "ljerns", "idngj"],
    "name": "Chad Jones"
  }
}
```

That's it! The power of JQ is that you can use it to target any field of any type, even searching across multiple objects for similar named fields and more. It's truly the most flexible way to transform your data.

### Triggering a Job Run

Another common use case is to create resources in Neosync such as Jobs, Connections, Runs, Transformers and more. In this example, we'll trigger a Job which will create a Job Run. This can be used as part of a set-up script or custom workflow. Let's take a look at the code:

Let's augment our code from above to call the `CreateJobRun` API.

```go
func main() {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		newHttpClient(map[string]string{ // create an instance of a newHttpClient
			"Authorization": fmt.Sprintf("Bearer %s", os.GetEnv("API_KEY")), // pass in the API_KEY through an environment variable
		}),
		os.GetEnv("API_URL"), // pass in the API_URL through an environment variable
	)

	// Calling the CreateJobRun in the JobServiceClient
	_, err := jobclient.CreateJobRun(context.Background(), connect.NewRequest(&mgmtv1alpha1.CreateJobRunRequest{
		JobId: "<job-id>",
	}))
	if err != nil {
		panic(err)
	}
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

## Moving forward

Now that you've seen how to anonymize data, generate synthetic data and create resources in Neosync, you can use the Neosync Go SDK to do much more! And if you have any questions, we're always available in Discord to help.
