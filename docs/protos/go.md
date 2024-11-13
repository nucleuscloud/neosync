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
    "favorites": ["dog", "cat", "bird"]
  }
}
```

Our input object is a simple user's object that we may get through a user sign up flow. In this object, we have a few sensitive fields that we want to anonymize: `name`, `email`, `address` and `phone`. We can leave the `favorites` as-is for now.

In order to anonymize this object, you can use Neosync's `AnonymizeSingle` API to send in a single object with sensitive data and get back an anonymized version of that object. You have full control over how you anonymize the data or generate new synthetic data.

Here's how you do it:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

// define our User struct and use json tags to structure our json
type User struct {
	User    UserDefinition `json:"user"`
	Details UserDetails    `json:"details"`
}

type UserDefinition struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserDetails struct {
	Address   string   `json:"address"`
	Phone     string   `json:"phone"`
	Favorites []string `json:"favorites"`
}

func main() {
	anonymizeClient := mgmtv1alpha1connect.NewAnonymizationServiceClient(
			newHttpClient(map[string]string{ // create an instance of a newHttpClient
			"Authorization": fmt.Sprintf("Bearer %s", os.GetEnv("API_KEY")), // pass in the API_KEY through an environment variable
		}),
		os.GetEnv("API_URL"), // pass in the API_URL through an environment variable
	)

	inputData := User{
		User: UserDefinition{
			Name:  "Bob Smith",
			Email: "random@different.com",
		},
		Details: UserDetails{
			Address: "123 Main St",
			Phone:   "555-1234",
			Favorites: []string{
				"cat", "dog", "cow",
			},
		},
	}

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

	// marshal our object into bytes
	userBytes, err := json.Marshal(inputData)
	if err != nil {
		panic(err)
	}

	resp, err := anonymizeClient.AnonymizeSingle(context.Background(), connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
		InputData:           string(userBytes), //stringify our bytes
		TransformerMappings: transformerMappings,
	}))
	if err != nil {
		fmt.Printf("Error in AnonymizeSingle: %v\n", err)
		panic(err)
	}

	fmt.Printf("Anonymization response: %+v\n", resp.Msg.OutputData)
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

Let's take a closer look at what we're doing here. Neosync's AnonymizeSingle API uses [JQ](https://jqlang.github.io/jq/manual/) expressions to target field(s) in your object. This means that you don't have to parse your object before sending it to Neosync. You can pass it in as-is and just write JQ expressions to target the field(s) that you want to anonymize or generate.

Our output will look something like this:

```json
// output
"{\"details\":{\"address\":\"123 Main St\",\"favorites\":[\"idh\",\"tyj\",\"ean\"],\"phone\":\"555-1234\"},\"user\":{\"email\":\"60ff1f2bb443484b928404164481f7f6@hootsuite.com\",\"name\":\"Nim Racic\"}}"
```

That's it! The power of JQ is that you can use it to target any field of any type, even searching across multiple objects for similar named fields and more. It's truly the most flexible way to transform your data.

### Anonymizing Unstructured Data

Another common use case is to anonymize free form text or unstructured data. This is useful in a variety of use-cases from doctor's notes to legal notes to chatbots and more.

The best part is that all you have to do is change a transformer, that's it! Here's how:

```js
// input
 {
    text: "Dear Mr. John Chang, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist Jake is on 8/1/2024 at 11 AM. Please bring a photo ID. We have your SSN on file as 246-80-1357. Is this correct?",
  },
```

Our input object is a transcription from a call from a doctor's office. In this transcript, we have PII (personally identifiable information) such as names (John Chang, Jake), social security number (246-80-1357) and dates(8/1/2024). Using Neosync's `TransformPiiText` transformer, you can easily anonymize the sensitive data in this text. See [here](https://docs.neosync.dev/api/mgmt/v1alpha1/transformer.proto#transformpiitext) for the `TransformPiiText` proto definition.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

// define our User struct and use json tags to structure our json
type Notes struct {
	Text   string `json:"text"`
}

func main() {
	anonymizeClient := mgmtv1alpha1connect.NewAnonymizationServiceClient(
			newHttpClient(map[string]string{ // create an instance of a newHttpClient
			"Authorization": fmt.Sprintf("Bearer %s", os.GetEnv("API_KEY")), // pass in the API_KEY through an environment variable
		}),
		os.GetEnv("API_URL"), // pass in the API_URL through an environment variable
	)

	inputData := Notes{
		Text: "Dear Mr. John Chang, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist Jake is on 8/1/2024 at 11 AM. Please bring a photo ID. We have your SSN on file as 246-80-1357. Is this correct?",
	}

	transformerMappings := []*mgmtv1alpha1.TransformerMapping{
		{
			Expression: `.notes.text`, // transform notes.text field
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPiiText{
					TransformPiiTextConfig: &mgmtv1alpha1.TransformPiiText{
						ScoreThreshold: 0.1 // lower = more paranoid, higher chance of false positive; higher = less paranoid, higher chance of false negative
					},
				},
			},
		},
	}

	// marshal our object into bytes
	notesBytes, err := json.Marshal(inputData)
	if err != nil {
		panic(err)
	}

	resp, err := anonymizeClient.AnonymizeSingle(context.Background(), connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
		InputData:           string(notesBytes), //stringify our bytes
		TransformerMappings: transformerMappings,
		AccountId: "xxxx". // your accountId found in the the App settings
	}))
	if err != nil {
		fmt.Printf("Error in AnonymizeSingle: %v\n", err)
		panic(err)
	}

	fmt.Printf("Anonymization response: %+v\n", resp.Msg.OutputData)
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

Let's take a closer look at what we're doing here. Neosync's AnonymizeSingle API uses [JQ](https://jqlang.github.io/jq/manual/) expressions to target field(s) in your object. This means that you don't have to parse your object before sending it to Neosync. You can pass it in as-is and just write JQ expressions to target the field(s) that you want to anonymize or generate.

Our output will look something like this:

```js
// output
Anonymization result: '{"text":"Dear Mr. \u003cREDACTED\u003e, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist \u003cREDACTED\u003e is on \u003cREDACTED\u003e at \u003cREDACTED\u003e. Please bring a photo ID. We have your SSN on file as \u003cREDACTED\u003e. Is this correct?"}'
```

As you can see, we've identified and redacted the PII in the original message and output a string that no longer contains PII. Alternatively, you can choose to Replace, Mask or even Hash the detected PII value instead of Redacting it.

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
