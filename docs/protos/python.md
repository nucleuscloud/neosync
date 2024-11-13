---
title: Neosync for Python
description: Learn about Neosync's Python SDK and how you can use it to anonymize data and generate synthetic data
id: python
hide_title: false
slug: /python
---

## Introduction

The Neosync Python SDK is publicly available and can be added to any Python project. With the Neosync Python SDK, you can:

1. Anonymize structured data and generate synthetic data
2. Anonymize free-form text data
3. Create resources in Neosync such as Jobs, Connections, Transformers and more

## Installation

You can add the Neosync Python SDK using:

`pip install neosync`

## Prerequisites

There are a few prerequisites that the SDK needs in order to be properly configured.

| **Properties** | **Details**                                                                                                                                                                 |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **API URL**    | Production: `https://neosync-api.svcs.neosync.dev`<br /> Local: `http://localhost:8080`                                                                                     |
| **Account ID** | The account ID may be necessary for some requests and can be found by going into the `/:accountName/settings` page in the Neosync App                                       |
| **API Key**    | An access token (API key, or user JWT) must be used to access authenticated Neosync environments. For an API Key, this can be created at `/:accountName/settings/api-keys`. |

## Authentication

If you are using Neosync locally and in unauthenticated mode then there is no authentication required and you can move onto the [Getting Started](go#getting-started) section.

If you are using Neosync locally in `auth mode` or using Neosync Cloud, you can authenticate with the Neosync server using an API URL and API Key. Here is an example showing how to authenticate with Neosync's GRPC server.

```python
import grpc
import json
import os

from neosync.mgmt.v1alpha1.anonymization_pb2_grpc import AnonymizationServiceStub
from neosync.mgmt.v1alpha1.anonymization_pb2 import AnonymizeSingleRequest,TransformerMapping
from neosync.mgmt.v1alpha1.transformer_pb2 import TransformerConfig, TransformPiiText

def main():

    # Creates TLS credentials for secure communication
    channel_credentials = grpc.ssl_channel_credentials()

    # Opens a secure channel to communication to the GRPC endpoint
    channel = grpc.secure_channel(os.getenv('NEOSYNC_API_URL'), channel_credentials)

    # Creates an client stub that we can use to invoke the AnonymizationService
    stub = AnonymizationServiceStub(channel)

    # Defines the Transformer that we want to use
    transformer_mapping = TransformerMapping(
        expression=".text",
        transformer=TransformerConfig(
            transform_pii_text_config=TransformPiiText(
                score_threshold=0.5,
            )
        )
    )

    # Defines our data input
    data = {'text': "This is John Williams and I need to be anonymized!"}

    # Stringifies the json in order to pass it into the API
    input_data = json.dumps(data)

    # Sets the Neosync API token to authenticate with the Neosync API Server
    access_token = '<neosnc-api-token>'

    # Defines metadata and an authorization header
    metadata = [('authorization', f'Bearer {access_token}')] if access_token else None

    try:
            # Make the RPC call
            response = stub.AnonymizeSingle(
                AnonymizeSingleRequest(
                    input_data=input_data,
                    transformer_mappings=[transformer_mapping],
                    account_id="<neosync-aaccount-id>"
                ),
                metadata=metadata
            )

            # Parse and print response
            try:
                output_data = json.loads(response.output_data)
                print("Anonymized data:", output_data)
            except json.JSONDecodeError:
                print("Raw response:", response.output_data)
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.UNAUTHENTICATED:
            print("Authentication failed - please check your access token")
        elif e.code() == grpc.StatusCode.UNAVAILABLE:
            print("Failed to connect to server. Please check the hostname and port.")
        else:
            print(f"RPC failed: {e.code()}, {e.details()}")
    finally:
        channel.close()
if __name__ == "__main__":
    main()
```

## Getting started

In this section, we're going to walk through two examples that show you how to make an API call using Neosync's Python SDK. For a complete list of the APIs, check out the APIs in the `Services` section of our [protos](/api/mgmt/v1alpha1/job.proto#jobservice).

Neosync is made up of a number of different services that live inside of the same process. All of the services are exposed through a stub class called `<service_name>Stub` which you can import from the `neosync.mgmt.v1alpa1.<service_name>_pb2_grpc` module. For example, the Anonymization service is available at:

- `from neosync.mgmt.v1alpha1.anonymization_pb2_grpc import AnonymizationServiceStub`

The types for this package are available at:

- `from neosync.mgmt.v1alpha1.anonymization_pb2 import AnonymizeSingleRequest`

Notice the different in the file names between the GRPC service stubs and the types. The GRPC service stubs end in a `_grpc` suffix while the types end in `_pb2`.

All services are structured this way.

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

```python

import grpc
import json
import os

from neosync.mgmt.v1alpha1.anonymization_pb2_grpc import AnonymizationServiceStub
from neosync.mgmt.v1alpha1.anonymization_pb2 import AnonymizeSingleRequest,TransformerMapping
from neosync.mgmt.v1alpha1.transformer_pb2 import TransformerConfig, TransformCharacterScramble, TransformFullName, TransformEmail

def main():

    channel_credentials = grpc.ssl_channel_credentials()

    channel = grpc.secure_channel(os.getenv('NEOSYNC_API_URL'), channel_credentials)
    stub = AnonymizationServiceStub(channel)

    transformer_mapping = [TransformerMapping(
        expression=".user.name",
        transformer=TransformerConfig(
            transform_full_name_config=TransformFullName(
                preserve_length=True,
            )
        )
    ),TransformerMapping(
        expression=".user.email",
        transformer=TransformerConfig(
            transform_email_config=TransformEmail(
            )
        )
    ),TransformerMapping(
        expression=".details.favorites[]",
        transformer=TransformerConfig(
            transform_character_scramble_config=TransformCharacterScramble(
            )
        )
    )]

    data = {
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

    input_data = json.dumps(data)

    access_token = 'neosync-api-token'
    metadata = [('authorization', f'Bearer {access_token}')] if access_token else None

    try:
            # Make the RPC call
            response = stub.AnonymizeSingle(
                AnonymizeSingleRequest(
                    input_data=input_data,
                    transformer_mappings=transformer_mapping,
                    account_id="neosync-account-id"
                ),
                metadata=metadata
            )

            # Parse and print response
            try:
                output_data = json.loads(response.output_data)
                print("Anonymized data:", output_data)
            except json.JSONDecodeError:
                print("Raw response:", response.output_data)
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.UNAUTHENTICATED:
            print("Authentication failed - please check your access token")
        elif e.code() == grpc.StatusCode.UNAVAILABLE:
            print("Failed to connect to server. Please check the hostname and port.")
        else:
            print(f"RPC failed: {e.code()}, {e.details()}")
    finally:
        channel.close()
if __name__ == "__main__":
    main()

```

Let's take a closer look at what we're doing here. Neosync's AnonymizeSingle API uses [JQ](https://jqlang.github.io/jq/manual/) expressions to target field(s) in your object. This means that you don't have to parse your object before sending it to Neosync. You can pass it in as-is and just write JQ expressions to target the field(s) that you want to anonymize or generate.

Our output will look something like this:

```json
// output
{
  "details": {
    "address": "123 Main St",
    "favorites": ["cce", "dya", "mqre"],
    "phone": "555-1234"
  },
  "user": {
    "email": "2cc88ac202624b3aa114694e57c1e206@justdial.com",
    "name": "Anab Bzt"
  }
}
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

```python
import grpc
import json
import os

from neosync.mgmt.v1alpha1.anonymization_pb2_grpc import AnonymizationServiceStub
from neosync.mgmt.v1alpha1.anonymization_pb2 import AnonymizeSingleRequest,TransformerMapping
from neosync.mgmt.v1alpha1.transformer_pb2 import TransformerConfig, TransformPiiText

def main():

    channel_credentials = grpc.ssl_channel_credentials()

    channel = grpc.secure_channel(os.getenv('NEOSYNC_API_URL'), channel_credentials)
    stub = AnonymizationServiceStub(channel)

    transformer_mapping = TransformerMapping(
        expression=".text",
        transformer=TransformerConfig(
            transform_pii_text_config=TransformPiiText(
                score_threshold=0.5,
            )
        )
    )

    data =  {"text": "Dear Mr. John Chang, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist Jake is on 8/1/2024 at 11 AM. Please bring a photo ID. We have your SSN on file as 246-80-1357. Is this correct?"}

    input_data = json.dumps(data)

    access_token = '<neosync-api-token>'
    metadata = [('authorization', f'Bearer {access_token}')] if access_token else None

    try:
            # Make the RPC call
            response = stub.AnonymizeSingle(
                AnonymizeSingleRequest(
                    input_data=input_data,
                    transformer_mappings=[transformer_mapping],
                    account_id="<neosync-account-id>"
                ),
                metadata=metadata
            )

            # Parse and print response
            try:
                output_data = json.loads(response.output_data)
                print("Anonymized data:", output_data)
            except json.JSONDecodeError:
                print("Raw response:", response.output_data)
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.UNAUTHENTICATED:
            print("Authentication failed - please check your access token")
        elif e.code() == grpc.StatusCode.UNAVAILABLE:
            print("Failed to connect to server. Please check the hostname and port.")
        else:
            print(f"RPC failed: {e.code()}, {e.details()}")
    finally:
        channel.close()
if __name__ == "__main__":
    main()
```

Let's take a closer look at what we're doing here. Neosync's AnonymizeSingle API uses [JQ](https://jqlang.github.io/jq/manual/) expressions to target field(s) in your object. This means that you don't have to parse your object before sending it to Neosync. You can pass it in as-is and just write JQ expressions to target the field(s) that you want to anonymize or generate.

Our output will look something like this:

```js
// output
{'text': 'Dear Mr. <REDACTED>, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist <REDACTED> is on <REDACTED> at <REDACTED>. Please bring a photo ID. We have your SSN on file as <REDACTED>. Is this correct?'}
```

As you can see, we've identified and redacted the PII in the original message and output a string that no longer contains PII. Alternatively, you can choose to Replace, Mask or even Hash the detected PII value instead of Redacting it.

### Triggering a Job Run

Another common use case is to create resources in Neosync such as Jobs, Connections, Runs, Transformers and more. In this example, we'll trigger a Job which will create a Job Run. This can be used as part of a set-up script or custom workflow. Let's take a look at the code:

Let's augment our code from above to call the `CreateJobRun` API.

```python
import grpc
import json
import os

from neosync.mgmt.v1alpha1.job_pb2_grpc import JobServiceStub
from neosync.mgmt.v1alpha1.job_pb2 import CreateJobRunRequest

from google.protobuf.json_format import MessageToDict, MessageToJson

def main():

    channel_credentials = grpc.ssl_channel_credentials()

    channel = grpc.secure_channel(os.getenv('NEOSYNC_API_URL'), channel_credentials)
    stub = JobServiceStub(channel)

    access_token = '<neosync-api-token>'
    metadata = [('authorization', f'Bearer {access_token}')] if access_token else None

    try:
            # Make the RPC call
            response = stub.CreateJobRun(
                CreateJobRunRequest(job_id="<job-id>"),
                metadata=metadata
            )
            # Parse and print response
            try:
                response_json = MessageToJson(response)
                print(response_json)
            except json.JSONDecodeError:
                print("Raw response:", response)
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.UNAUTHENTICATED:
            print("Authentication failed - please check your access token")
        elif e.code() == grpc.StatusCode.UNAVAILABLE:
            print("Failed to connect to server. Please check the hostname and port.")
        else:
            print(f"RPC failed: {e.code()}, {e.details()}")
    finally:
        channel.close()
if __name__ == "__main__":
    main()
```

## Moving forward

Now that you've seen how to anonymize data, generate synthetic data and create resources in Neosync, you can use the Neosync Python SDK to do much more! And if you have any questions, we're always available in Discord to help.
