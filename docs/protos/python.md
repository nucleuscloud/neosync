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

| **Properties** | **Details**                                                                                                                                                                 | **Default**                      |
|----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------|
| **Account ID** | The account ID may be necessary for some requests and can be found by going into the `/:accountName/settings` page in the Neosync App.                                      |                                  |
| **API Key**    | An access token (API key, or user JWT) must be used to access authenticated Neosync environments. For an API Key, this can be created at `/:accountName/settings/api-keys`. |                                  |
| **API URL**    | The instance of Neosync to point to.                                                                                                                                         | neosync-api.svcs.neosync.dev:443 |

## Authentication

If you are using Neosync locally and in unauthenticated mode then there is no authentication required and you can move onto the [Getting Started](go#getting-started) section.

If you are using Neosync locally in `auth mode` or using **Neosync Cloud**, you can authenticate with the Neosync server using an API Key or Access Token. Here is an example showing how to authenticate with Neosync's server.

```python
from neosync import Neosync
import os

def main():
    client = Neosync(
        # Optional, defaults to Neosync Cloud if not provided
        api_url=os.getenv("NEOSYNC_API_URL"),
        # Optional if connecting to a custom instance of Neosync with no auth
        access_token=os.getenv("NEOSYNC_API_KEY"),
    )

if __name__ == "__main__":
    main()
```

## Getting started

In this section, we're going to walk through two examples that show you how to make an API call using Neosync's Python SDK. For a complete list of the APIs, check out the APIs in the `Services` section of each proto file. ex.: [AnonymizationService](/api/mgmt/v1alpha1/anonymization.proto#anonymizationservice).

Neosync is made up of a number of different services that live inside of the same process. These are all present on the `Neosync` client class.

Each function that is called has a corresponding request and response class.

For example, the `AnonymizeSingle` function has a corresponding request class called `AnonymizeSingleRequest` and a corresponding response class called `AnonymizeSingleResponse`.

These are all importable directly from the root `neosync` package, or may be imported from the `neosync.mgmt.v1alpha1` package.

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

import json
import os

from neosync import (
    Neosync,
    AnonymizeSingleRequest,
    AnonymizeSingleResponse,
    TransformerMapping,
    TransformerConfig,
    TransformCharacterScramble,
    TransformFullName,
    TransformEmail,
)

ACCESS_TOKEN = os.getenv("NEOSYNC_API_KEY")
ACCOUNT_ID = os.getenv("NEOSYNC_ACCOUNT_ID")


def main():
    client = Neosync(access_token=ACCESS_TOKEN)

    input_data = json.dumps(
        {
            "user": {"name": "John Doe", "email": "john@example.com"},
            "details": {
                "address": "123 Main St",
                "phone": "555-1234",
                "favorites": ["dog", "cat", "bird"],
            },
        }
    )

    response: AnonymizeSingleResponse = client.anonymization.AnonymizeSingle(
        AnonymizeSingleRequest(
            input_data=input_data,
            transformer_mappings=[
                TransformerMapping(
                    expression=".user.name",
                    transformer=TransformerConfig(
                        transform_full_name_config=TransformFullName(
                            preserve_length=True,
                        )
                    ),
                ),
                TransformerMapping(
                    expression=".user.email",
                    transformer=TransformerConfig(
                        transform_email_config=TransformEmail()
                    ),
                ),
                TransformerMapping(
                    expression=".details.favorites[]",
                    transformer=TransformerConfig(
                        transform_character_scramble_config=TransformCharacterScramble()
                    ),
                ),
            ],
            account_id=ACCOUNT_ID,
        ),
    )
    print(json.loads(response.output_data))


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
import json
import os

from neosync import (
    Neosync,
    AnonymizeSingleRequest,
    AnonymizeSingleResponse,
    TransformerMapping,
    TransformerConfig,
    TransformPiiText,
)

ACCESS_TOKEN = os.getenv("NEOSYNC_API_KEY")
ACCOUNT_ID = os.getenv("NEOSYNC_ACCOUNT_ID")


def main():
    client = Neosync(access_token=ACCESS_TOKEN)

    input_data = json.dumps(
        {
            "text": "Dear Mr. John Chang, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist Jake is on 8/1/2024 at 11 AM. Please bring a photo ID. We have your SSN on file as 246-80-1357. Is this correct?"
        }
    )

    response: AnonymizeSingleResponse = client.anonymization.AnonymizeSingle(
        AnonymizeSingleRequest(
            input_data=input_data,
            transformer_mappings=[
                TransformerMapping(
                    expression=".text",
                    transformer=TransformerConfig(
                        transform_pii_text_config=TransformPiiText(
                            score_threshold=0.5,
                        )
                    ),
                )
            ],
            account_id=ACCOUNT_ID,
        ),
    )
    print(json.loads(response.output_data))


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
import os

from neosync import (
    Neosync,
    CreateJobRunRequest,
)

ACCESS_TOKEN = os.getenv("NEOSYNC_API_KEY")


def main():
    client = Neosync(access_token=ACCESS_TOKEN)

    # Create a job run, returns an empty response
    client.jobs.CreateJobRun(
        CreateJobRunRequest(job_id="your-job-id"),
    )


if __name__ == "__main__":
    main()

```

## Error Handling

Most of these examples show the happy path. This section details how to properly catch errors when using the Python SDK.

Today the Python SDK connects to Neosync using gRPC. This means that you can catch errors using the `grpc` library, which is installed as a dependency to this SDK.

Here is the anonymize single example from above, but with some basic error handling added in.

```python
import json
import os
import grpc

from neosync import (
    Neosync,
    AnonymizeSingleRequest,
    AnonymizeSingleResponse,
    TransformerMapping,
    TransformerConfig,
    TransformPiiText,
)

ACCESS_TOKEN = os.getenv("NEOSYNC_API_KEY")
ACCOUNT_ID = os.getenv("NEOSYNC_ACCOUNT_ID")


def main():
    client = Neosync(access_token=ACCESS_TOKEN)

    input_data = json.dumps(
        {
            "text": "Dear Mr. John Chang, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist Jake is on 8/1/2024 at 11 AM. Please bring a photo ID. We have your SSN on file as 246-80-1357. Is this correct?"
        }
    )

    try:
        response: AnonymizeSingleResponse = client.anonymization.AnonymizeSingle(
            AnonymizeSingleRequest(
                input_data=input_data,
                transformer_mappings=[
                    TransformerMapping(
                        expression=".text",
                        transformer=TransformerConfig(
                            transform_pii_text_config=TransformPiiText(
                                score_threshold=0.5,
                            )
                        ),
                    )
                ],
                account_id=ACCOUNT_ID,
            ),
        )
        print(json.loads(gresponse.output_data))
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.UNAUTHENTICATED:
            print("Authentication failed - please check your access token")
        elif e.code() == grpc.StatusCode.UNAVAILABLE:
            print("Failed to connect to server. Please check the hostname and port.")
        else:
            print(f"RPC failed: {e.code()}, {e.details()}")

if __name__ == "__main__":
    main()

```

## Decoding Anonymization `output_data`

When using the `TransformPiiText` transformer, you may notice that the output output of redacted text can sometimes be in an escaped format. So instead of `<PERSON>` you see `\u003cPERSON\u003e`. This is because the `output_data` is returned as a stringifed JSON object.

There are a few ways to decode this properly:

### Decoding as a JSON object

The `output_data` is a stringified JSON object. You can use `json.loads()` to decode the string into a JSON object.

### Decoding as a string

You can use the `response.output_data.encode().decode("unicode-escape")` method.

This will properly decode the escaped characters in the output data. This can be printed or passed into `json.loads()` to then extract your data as needed.

## Context Manager

The Neosync Python SDK also supports context managers. This means that you can use the `with` statement to create a Neosync client and automatically close the gRPC channel when you're done.
Otherwise, you'll need to call `client.close()` manually (only if it's desired to cleanly close the gRPC channel).

Here's an example:

```python
from neosync import Neosync

def main():
    with Neosync() as client:
        # Do something with the client
        pass

if __name__ == "__main__":
    main()
```

## Moving forward

Now that you've seen how to anonymize data, generate synthetic data and create resources in Neosync, you can use the Neosync Python SDK to do much more! And if you have any questions, we're always available in Discord to help.
