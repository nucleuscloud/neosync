---
title: Anonymization Service Endpoints
description: Guide to Using the Anonymization Service Endpoints
id: anonymization-service-endpoints
hide_title: false
slug: /guides/anonymization-service-endpoints
# cSpell:words Neosync,Protos,cmedicine
---

## Introduction

This guide provides an overview of how to use the `AnonymizeSingle`and `AnonymizeMany` endpoints of the anonymization service. These endpoints are designed to anonymize JSON data by applying specified transformation rules using [JQ (JSON Query)](https://jqlang.github.io/jq/manual/) paths and expressions.

> See the [Anonymization Service Protos](/api/mgmt/v1alpha1/user_account.proto) page for more detailed information.

> See the [TS SDK](/api/typescript) or [GO SDK](/api/go) page for more information on how to use the SDK.

> [JQ Playground](https://jqplay.org/s/slj84ii6GZbbaw7)

### JQ Paths vs JQ Expressions

**JQ Paths**:

- Example: `.path.to.field`
- These are specific paths to fields in the JSON structure.
- They always begin with .
- Use ? (optional operator) for optional fields. It suppresses errors that would otherwise be raised if a certain key, field, or expression doesn't exist. Example .departments[]?.projects[]?.name
- Transformer mappings using JQ paths take precedence over default transformers.

**JSON Array Support in JQ Paths**:

- For JSON arrays, only indices are supported (e.g., [0], [1], [2]).
- The [] notation is supported to select all elements of an array.
- Slice notation (e.g., [1:4]) is not supported.

**JQ Expressions**:

- Example: `(.. | objects | select(has("name")) | .name)`
- These are more complex queries that can select multiple fields across the JSON structure.
- Expressions should always be in parenthesis ()
- Important: Default transformers will be applied first and then JQ expression transformers.

## AnonymizeSingle, AnonymizeMany

### Input Data & Output Data

`AnonymizeSingle`: Single JSON string

`AnonymizeMany`: Array of JSON strings

### Transformer Mappings (Required)

Transformer mappings define specific paths in your JSON data using JQ and the transformers to apply to those paths.

<!-- cspell:disable  -->

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

<!-- cspell:enable  -->

### Default Mappings (Optional)

Default transformers provide a way to anonymize fields that are not explicitly defined in the transformer mappings. They work on a data type basis, applying to all fields of a specific type that don't have a transformer mapping configured. It supports default transformers for three data types: strings, numbers, and booleans.

Specific transformer mappings always take precedence over default transformers. If a field has a transformer defined in the mappings using a JQ path, the default transformer for that data type will not be applied to that field.

This approach allows you to have a catch-all anonymization strategy for each data type while still maintaining fine-grained control where needed.

### JQ Expressions and Default Transformers

When using a JQ expression in your transformer mappings, be aware that the default transformers will still apply to the values found by the expression. It will apply the default transformer first and then apply the JQ expression transformer second. This behavior differs from using specific JQ paths where only the transformer configured for that path will be applied.

For example:

<!-- cspell:disable  -->

```go
transformerMappings := []*mgmtv1alpha1.TransformerMapping{
  {
    Expression: `(.. | objects | select(has("name")) | .name)`,
    Transformer: &mgmtv1alpha1.TransformerConfig{
      Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
        TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
          PreserveLength: true,
        },
      },
    },
  },
}

defaultTransformers := &mgmtv1alpha1.DefaultTransformersConfig{
  S: &mgmtv1alpha1.TransformerConfig{
    Config: &mgmtv1alpha1.TransformerConfig_GenerateRandomStringConfig{
      GenerateRandomStringConfig: &mgmtv1alpha1.GenerateRandomString{},
    },
  },
}
```

<!-- cspell:enable  -->

In this scenario:

1. The default string transformer will be applied to all string fields including name. The name field value will now be a random string.
2. The JQ expression `.. | objects | select(has("name")) | .name` will then find all "name" fields in the JSON and apply the full name transformer. This means the value passed into the full name transformer will be the random string from the string default transformer and not the original value.

### Best Practices

1. When you need to apply a specific transformation to fields selected by a JQ expression, consider disabling or not setting default transformers for that data type.
2. If you must use both JQ expressions and default transformers, carefully review the results to ensure the desired anonymization is achieved.
3. If using generators and not transformers then using both JQ expression and default transformers have no side effects.
4. For critical fields that need specific anonymization, prefer using exact JQ paths over JQ expressions when possible.

## AnonymizeMany

### Halt On Failure (Optional)

Determines the behavior of the anonymization process when encountering errors. Only successfully anonymized JSON strings are returned.

`true`: The process halts immediately upon encountering the first error.

`false`: The process continues anonymizing remaining data, collecting all errors.

## Customizing Entities

The Anonymization endpoint uses a Named Entity Recognition model under the covers to detect PII and then redact it. It comes with a base list of entities that it will automatically detect but you can also create your own entities.

### Retrieving the list of base entities

If you just want to use the base entities then you can leave the entities field blank. If you want to customize them, then you can see base list of entities that we support by calling the `getTransformPiiEntities` endpoint like this:

```typescript
const neosyncClient = getNeosyncClient({
  getAccessToken: () => {
    return api_key; // this is your api_key
  },
  getTransport(interceptors) {
    return createConnectTransport({
      baseUrl: api_url, // the api_url - check the SDK docs for the appropriate URL
      httpVersion: '2',
      interceptors: interceptors,
    });
  },
});

const entities = await neosyncClient.transformers.getTransformPiiEntities({
  accountId: accountId, // this is your account Id
});
```

This will return all of the supported entities.

### Filtering Entities

There are some situations where you do not want to detect certain PII entities. In these cases, you can filter out certain entities from being detected.

In the following example, we will filter out the `DATE_TIME` entity from the base list of entities.

```typescript
// client initialization

// ...

const entities = await neosyncClient.transformers.getTransformPiiEntities({
  accountId: accountId,
});

const filteredEntities = entities.entities.filter(
  (entity) => entity != 'DATE_TIME'
);

const transformers: TransformerMapping[] = [
  new TransformerMapping({
    expression: '.text',
    transformer: new TransformerConfig({
      config: {
        case: 'transformPiiTextConfig',
        value: new TransformPiiText({
          scoreThreshold: 0.2,
          allowedEntities: filteredEntities,
        }),
      },
    }),
  }),
];

// calling the anonymizeSingle endpoint

//...
```

This will detect every base entity beside the `DATE_TIME` entity.

### Adding Custom Entities

If you want to add Custom Entities that will always be redacted, you can create `denyRecognizers`.

In the following example, we're going to add `Advil` to `denyRecognizer` to ensure it's always redacted.

**Important**
If you are filtering the base list of entities at all as in the previous example, then you need to add the name of the `denyRecognizer`, in this case `medicine`, to the list of entities that you are passing into the `entities` field or else it won't be recognizer. If you are not changing the list of base entities at all, then you do not need to add it.

```typescript
// client initialization

// ...

//input with PII (Name, Medicine, Phone Number)
const data = {
  text: "Hello, yes, this is John Williams. I had an appointment 5 days ago to get some Advil and Ibuprofen, but didn't pick it up. Can you give me a call back at 6173943902? Thank you.",
};

// transformer configuration
const transformers: TransformerMapping[] = [
      new TransformerMapping({
        expression: ".text",
        transformer: new TransformerConfig({
          config: {
            case: "transformPiiTextConfig",
            value: new TransformPiiText({
              scoreThreshold: 0.2,
              denyRecognizers: [
                {
                  name: "medicine",
                  denyWords: ["Advil"],
                },
              ],
            }),
          },
        }),
      }),
    ];

// calling the anonymizeSingle endpoint

//...

// redacted output with the allowed phrases not redacted even though it's clearly a phone number
{"text":"Hello, yes, this is \u003cPERSON\u003e. I had an appointment \u003cDATE_TIME\u003e to get some \u003cmedicine\u003e and Ibuprofen, but didn't pick it up. Can you give me a call back at \u003cPHONE_NUMBER\u003e? Thank you."}
```

## Adding Allowed Phrases

The Allow Phrases list allows you to pass in a set of strings that will never get anonymized even if they are detected as PII. This is helpful if there are certain words that you need to always be in plain-text.

You can set the allow list by passing in an array of strings. For example:

```typescript
// client initialization

// ...


//input with PII (Name, Medicine, Phone Number)
const data = {
  text: "Hello, yes, this is John Williams. I had an appointment 5 days ago to get some Advil and Ibuprofen, but didn't pick it up. Can you give me a call back at 6173943902? Thank you.",
};

// transformer configuration
const transformers: TransformerMapping[] = [
      new TransformerMapping({
        expression: ".text",
        transformer: new TransformerConfig({
          config: {
            case: "transformPiiTextConfig",
            value: new TransformPiiText({
              scoreThreshold: 0.2,
              allowedPhrases: ["6173943902"],
            }),
          },
        }),
      }),
    ];

// calling the anonymizeSingle endpoint

//...

// redacted output with the allowed phrases not redacted even though it's clearly a phone number
{"text":"Hello, yes, this is \u003cPERSON\u003e. I had an appointment \u003cDATE_TIME\u003e to get some Advil and Ibuprofen, but didn't pick it up. Can you give me a call back at 6173943902? Thank you."}
```

## Error Handling

Errors can occur during the anonymization process due to malformed JSON, transformer failures, or other issues.

### AnonymizeSingle

If an error occurs, the response will contain an error message, and the output_data will be absent.

### AnonymizeMany

The response includes an errors field, which is a list of AnonymizeManyErrors objects. Each error corresponds to a specific input index and contains an error message.

If `halt_on_failure` is true, the output_data may contain partially anonymized data up to the point of failure.

If `halt_on_failure` is false, all possible data is anonymized, and errors are collected for problematic entries.
