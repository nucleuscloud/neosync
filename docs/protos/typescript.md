---
title: Neosync for TypeScript
description: Learn about Neosync's Typescript SDK and how you can use it to anonymize data and generate synthetic data
id: typescript
hide_title: false
slug: /typescript
---

## Introduction

The Neosync Typescript SDK is publicly available and can be added to any TS/JS-based project. With the Neosync Typescript SDK, you can:

1. Anonymize structured data and generate synthetic data
2. Anonymize free-form text data
3. Create resources in Neosync such as Jobs, Connections, Transformers and more

## Installation

This package supports both ES-Modules and CommonJS.

The correct entry point will be chosen based on using `import` or `require`.

The `tsup` package is used to generated the distributed code.

`@bufbuild/protobuf` provides methods to instantiate the messages used in the SDK.

```sh
npm install @neosync/sdk @bufbuild/protobuf
```

| **Properties** | **Details**                                                                                                                                                                 |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **API URL**    | Production: `https://neosync-api.svcs.neosync.dev`<br /> Local: `http://localhost:8080`                                                                                     |
| **Account ID** | The account ID may be necessary for some requests and can be found by going into the `/:accountName/settings` page in the Neosync App                                       |
| **API Key**    | An access token (API key, or user JWT) must be used to access authenticated Neosync environments. For an API Key, this can be created at `/:accountName/settings/api-keys`. |

### Note on Transports

Based on your usage, you'll have to install a different version of `connect` to provide the correct Transport based on your environment.

- Node: [@connectrpc/connect-node](https://connectrpc.com/docs/node/using-clients)
- Web: [@connectrpc/connect-web](https://connectrpc.com/docs/web/using-clients)

Install whichever one makes sense for you

```sh
npm install @connectrpc/connect-node
npm install @connectrpc/connect-web
```

## Authentication

To authenticate the TS Neosync Client, pass in a function that returns the API Key or a standard user JWT token. When the `getAccessToken` function is provided, the Neosync Client is configured with an auth interceptor that attaches the `Authorization` header to every outgoing request with the access token returned from the function. This is why the `getTransport` method receives a list of interceptors, and why it's important to hook them up to pass them through to the relevant transport being used.

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

## Getting Started

In this section, we're going to walk through two examples that show you how to make an API call using Neosync's TS SDK. For a complete list of the APIs, check out the APIs in the `Services` section of our [protos](/api/mgmt/v1alpha1/job.proto#jobservice).

### Note on Types and Messages

In each example there are some cases where the `create` function is used from the `@bufbuild/protobuf` package.

This is a convenience function that allows you to create a message from a schema.

It is generally only necessary for any top-level message that you are attempting to assign directly to any inferface.
The generaly pattern is this: `TransformerConfig` -> `TransformerConfigSchema` -> `create(TransformerConfigSchema, {})`.

The second parameter to the `create` function is a type that looks like this `MessageInit<TransformerConfig>`. which is effectively a `Partial<TransformerConfig>`.
This is the same interface that will be found on all of the actual RPC calls from the API. So if you are just inlining the messages directlyinto the RPC call, the `create` function is generally not necessary. The examples below highlight when to use the `create` function.

If inspecting the types or using your IDE's intellisense, you'll find that each message also contains two additional properties: `$typename` and `$unknown`.
These should not be set directly and are generally set by the `create` function. This information is used by the underlying library to ensure correct serialization and deserialization of the message.

### Anonymizing Structured Data

A straightforward use case is to anonymize sensitive data in an API request. Let's look at an example.

```jsonc
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
  }
}
```

Our input object is a simple user's object that we may get through a user sign up flow. In this object, we have a few sensitive fields that we want to anonymize: `name`, `email`, `address` and `phone`. We can leave the `favorites` as-is for now.

In order to anonymize this object, you can use Neosync's `AnonymizeSingle` API to send in a single object with sensitive data and get back an anonymized version of that object. You have full control over how you anonymize the data or generate new synthetic data.

Here's how you do it:

```ts
import { create } from '@bufbuild/protobuf';
import { createConnectTransport } from '@connectrpc/connect-node';
import {
  AnonymizeSingleResponse,
  getNeosyncClient,
  TransformerMapping,
  TransformerMappingSchema,
} from '@neosync/sdk';

// authenticates with Neosync Cloud
const neosyncClient = getNeosyncClient({
  getAccessToken: () => {
    return 'neo_at_v1_xxxxxxxxxxxx'; // API key
  },
  getTransport(interceptors) {
    return createConnectTransport({
      baseUrl: 'https://neosync-api.svcs.neosync.dev', // base url
      httpVersion: '2',
      interceptors: interceptors, // interceptors
    });
  },
});

// sample data object
const data = {
  user: {
    name: 'John Doe',
    email: 'john@example.com',
  },
  details: {
    address: '123 Main St',
    phone: '555-1234',
    favorites: ['dog', 'cat', 'bird'],
  },
};

const transformers: TransformerMapping[] = [
  create(TransformerMappingSchema, {
    expression: '.user.name', // targets the name field in the user object with a jq expression
    transformer: {
      config: {
        case: 'generateFullNameConfig', // sets the generateFullNameConfig
        value: {}, // sets the GenerateFullName transformer
      },
    },
  }),
  create(TransformerMappingSchema, {
    expression: '.user.email', // targets the email field in the user object with a jq expression
    transformer: {
      config: {
        case: 'generateEmailConfig', // sets the generateEmailConfig
        value: {}, // sets the GenerateEmail transformer
      },
    },
  }),
  create(TransformerMappingSchema, {
    expression: '.details.address', // targets the address field in the details object with a jq expression
    transformer: {
      config: {
        case: 'generateFullAddressConfig', // sets the generateFullAddressConfig
        value: {}, // sets the GenerateFullAddress transformer
      },
    },
  }),
  create(TransformerMappingSchema, {
    expression: '.details.phone', // targets the phone field in the details object with a jq expression
    transformer: {
      config: {
        case: 'generateStringPhoneNumberConfig', // sets the generateStringPhoneNumberConfig
        value: {
          // sets the GenerateStringPhoneNumber transformer
          max: BigInt(12), // sets the max number of digits in the string phone number
          min: BigInt(9), // sets the min number of digits in the string phone number
        },
      },
    },
  }),
];

async function runAnonymization() {
  try {
    const result: AnonymizeSingleResponse =
      await neosyncClient.anonymization.anonymizeSingle({
        inputData: JSON.stringify(data), // stringify the data object from above
        transformerMappings: transformers, // pass in your transformer mappings that you defined
      });
    console.log('Anonymization result:', result.outputData);
  } catch (error) {
    console.error('Error:', error);
  }
}

// calling our async function
runAnonymization()
  .then(() => console.log('Script completed'))
  .catch((error) => console.error('Unhandled error:', error));
```

Let's take a closer look at what we're doing here. Neosync's AnonymizeSingle API uses [JQ](https://jqlang.github.io/jq/manual/) expressions to target field(s) in your object. This means that you don't have to parse your object before sending it to Neosync. You can pass it in as-is and just write JQ expressions to target the field(s) that you want to anonymize or generate.

Our output will look something like this:

```js
// output result
Anonymization result: '{"user":{"email":"22fdd05dd75746728a9c2a37d3d58cf5@stackoverflow.com","name":"Bryam Begg"},"details":{"address":"212 Ambleside Drive Severna Park MD, 21146","favorites":["dog","cat","bird"],"phone":"58868075625"},}'
```

That's it! The power of JQ is that you can use it to target any field of any type, even searching across multiple objects for similar named fields and more. It's truly the most flexible way to transform your data.

### Anonymizing Unstructured Data

Another common use case is to anonymize free form text or unstructured data. This is useful in a variety of use-cases from doctor's notes to legal notes to chatbots and more.

The best part is that all you have to do is change a transformer, that's it! Here's how:

```jsonc
// input
 {
    text: "Dear Mr. John Chang, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist Jake is on 8/1/2024 at 11 AM. Please bring a photo ID. We have your SSN on file as 246-80-1357. Is this correct?",
  },
```

Our input object is a transcription from a call from a doctor's office. In this transcript, we have PII (personally identifiable information) such as names (John Chang, Jake), social security number (246-80-1357) and dates(8/1/2024). Using Neosync's `TransformPiiText` transformer, you can easily anonymize the sensitive data in this text. See [here](https://docs.neosync.dev/api/mgmt/v1alpha1/transformer.proto#transformpiitext) for the `TransformPiiText` proto definition.

```ts
import {
  AnonymizeSingleResponse,
  getNeosyncClient,
  TransformerMapping,
  TransformerMappingSchema,
} from '@neosync/sdk';
import { create } from '@bufbuild/protobuf;
import { createConnectTransport } from '@connectrpc/connect-node';

const neosyncClient = getNeosyncClient({
  getAccessToken: () => {
    return 'neo_at_v1_xxxxx'; // Neosync API Key
  },
  getTransport(interceptors) {
    return createConnectTransport({
      baseUrl: 'https://neosync-api.svcs.neosync.dev', // Neosync API Url
      httpVersion: '2',
      interceptors: interceptors,
    });
  },
});

const data = {
  text: 'Dear Mr. John Chang, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist Jake is on 8/1/2024 at 11 AM. Please bring a photo ID. We have your SSN on file as 246-80-1357. Is this correct?',
};

const transformers: TransformerMapping[] = [
  create(TransformerMappingSchema, {
    expression: '.text',
    transformer: {
      config: {
        case: 'transformPiiTextConfig', // set the case to transformPiiTextConfig
        value: {
          // use the TransformPiiText transformer
          scoreThreshold: 0.1, // lower = more paranoid, higher chance of false positive; higher = less paranoid, higher chance of false negative
        },
      },
    },
  }),
];

// calling our async function
runAnonymization()
  .then(() => console.log('Script completed'))
  .catch((error) => console.error('Unhandled error:', error));
```

Let's take a closer look at what we're doing here. Neosync's AnonymizeSingle API uses [JQ](https://jqlang.github.io/jq/manual/) expressions to target field(s) in your object. This means that you don't have to parse your object before sending it to Neosync. You can pass it in as-is and just write JQ expressions to target the field(s) that you want to anonymize or generate.

Our output will look something like this:

```js
// output
Anonymization result: '{"text":"Dear Mr. \u003cREDACTED\u003e, your physical therapy for your rotator cuff injury is approved for 12 sessions. Your first appointment with therapist \u003cREDACTED\u003e is on \u003cREDACTED\u003e at \u003cREDACTED\u003e. Please bring a photo ID. We have your SSN on file as \u003cREDACTED\u003e. Is this correct?"}'
```

As you can see, we've identified and redacted the PII in the original message and output a string that no longer contains PII. Alternatively, you can choose to Replace, Mask or even Hash the detected PII value instead of Redacting it.

### Creating a Job

Another common use case is to create resources in Neosync such as Jobs, Runs, Connections, Transformers and more. In this example, we'll create a Job. This can be used as part of a set-up script or custom workflow. Let's take a look at the code:

```ts
import { createConnectTransport } from '@connectrpc/connect-node';
import { create } from '@bufbuild/protobuf';
import {
  CreateJobDestination,
  CreateJobDestinationSchema,
  CreateJobResponse,
  getNeosyncClient,
  JobMapping,
  JobMappingSchema,
  JobSource,
  JobSourceSchema,
} from '@neosync/sdk';

// authenticates with Neosync Cloud
const neosyncClient = getNeosyncClient({
  getAccessToken: () => {
    return 'neo_at_v1_xxxxxxxxxxxx'; // API key
  },
  getTransport(interceptors) {
    return createConnectTransport({
      baseUrl: 'https://neosync-api.svcs.neosync.dev', // base url
      httpVersion: '2',
      interceptors: interceptors, // interceptors
    });
  },
});

// creates our job mappings which maps transformers -> columns
const jobMapping: JobMapping[] = [
  create(JobMappingSchema, {
    schema: 'public',
    table: 'users',
    column: 'email', // mapping the email column
    transformer: {
      config: {
        config: {
          case: 'generateEmailConfig', // setting the generateEmailConfig
          value: {}, // setting the GenerateEmail transformer to the email column
        },
      },
    },
  }),
  create(JobMappingSchema, {
    schema: 'public',
    table: 'users',
    column: 'age', // mapping the age column
    transformer: {
      config: {
        config: {
          case: 'generateInt64Config', // setting the generateInt64Config
          value: {}, // setting the GenerateInt64 transformer to the age column
        },
      },
    },
  }),
  create(JobMappingSchema, {
    schema: 'public',
    table: 'users',
    column: 'address', // mapping the address column
    transformer: {
      config: {
        config: {
          case: 'generateFullAddressConfig', // setting the generateFullAddressConfig
          value: {}, // setting the GenerateFullAddress transformer to the address column
        },
      },
    },
  }),
];

// setting our source connection and connection optinos
const sourceConnection: JobSource = create(JobSourceSchema, {
  options: {
    config: {
      case: 'postgres',
      value: {
        schemas: [],
        connectionId: '4efaff59-ed4d-4365-8e0e-eccad4a49481',
        subsetByForeignKeyConstraints: false,
        haltOnNewColumnAddition: false,
      },
    },
  },
});

// setting our destination
const destination: CreateJobDestination[] = [
  create(CreateJobDestinationSchema, {
    connectionId: '3470533a-1fcc-43ec-9cba-8c037ea0da47',
  }),
];

async function createJob() {
  try {
    // calling the jobs.createJobs rpc with our configurations in order to create a job called 'new-job'
    const result: CreateJobResponse = await neosyncClient.jobs.createJob({
      accountId: 'b1b8411f-a2f5-4ca1-b710-2fbc2681527e',
      jobName: 'new-job',
      mappings: jobMapping,
      cronSchedule: '0 0 1 1 *',
      source: sourceConnection,
      destinations: destination,
      syncOptions: {},
    });
    // returning the job.id here
    console.log('Job result:', result.job?.id);
  } catch (error) {
    console.error('Error:', error);
  }
}

// calling our async function
createJob()
  .then(() => console.log('Script completed'))
  .catch((error) => console.error('Unhandled error:', error));
```

The beauty of Typescript here is that you can use your IDE's built-in features to see exactly what is required and what is optional. And if your IDE doesn't support that then you can use the protobuf files to see how the messages are constructed.

## Moving forward

Now that you've seen how to anonymize data, generate synthetic data and create resources in Neosync, you can use the Neosync TS SDK to do much more! And if you have any questions, we're always available in Discord to help.
