---
title: Anonymization Service Endpoints 
description: Guide to Using the Anonymization Service Endpoints
id: anonymization-service-endpoints
hide_title: false
slug: /guides/anonymization-service-endpoints
# cSpell:words Neosync
---

## Introduction

This guide provides an overview of how to use the `AnonymizeSingle`and `AnonymizeMany` endpoints of the anonymization service. These endpoints are designed to anonymize JSON data by applying specified transformation rules using [JQ (JSON Query)](https://jqlang.github.io/jq/manual/) paths and expressions.

> See the [Anonymization Service Protos](/api/mgmt/v1alpha1/user_account.proto) page for more detailed information.

> See the [TS SDK](/api/typescript) or [GO SDK](/api/go) page for more information on how to use the SDK.

### JQ Paths vs JQ Expressions

**JQ Paths**: 
   - Example: `.path.to.field`
   - These are specific paths to fields in the JSON structure.
   - Transformer mappings using JQ paths take precedence over default transformers.

**JQ Expressions**:
   - Example: `.. | objects | select(has("name")) | .name`
   - These are more complex queries that can select multiple fields across the JSON structure.
   - Important: Default transformers will be applied first and then JQ expression transformers.



## AnonymizeSingle, AnonymizeMany

### Transformer Mappings
Transformer mappings define specific paths in your JSON data using JQ and the transformers to apply to those paths.

### Default Mappings
Default transformers provide a way to anonymize fields that are not explicitly defined in the transformer mappings. They work on a data type basis, applying to all fields of a specific type that don't have a transformer mapping configured. It supports default transformers for three data types: strings, numbers, and booleans.

Specific transformer mappings always take precedence over default transformers. If a field has a transformer defined in the mappings using a JQ path, the default transformer for that data type will not be applied to that field.

 This approach allows you to have a catch-all anonymization strategy for each data type while still maintaining fine-grained control where needed.

### JQ Expressions and Default Transformers

When using a JQ expression in your transformer mappings, be aware that the default transformers will still apply to the values found by the expression. It will apply the default transformer first and then apply the JQ expression transformer second. This behavior differs from using specific JQ paths where only the transformer configured for that path will be applied.

For example:
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

In this scenario:
1. The default string transformer will be applied to all string fields including name. The name field value will now be a random string.
2. The JQ expression `.. | objects | select(has("name")) | .name` will then find all "name" fields in the JSON and apply the full name transformer. This means the value passed into the full name transformer will be the random string from the string default transformer and not the original value.

### Best Practices

1. When you need to apply a specific transformation to fields selected by a JQ expression, consider disabling or not setting default transformers for that data type.
2. If you must use both JQ expressions and default transformers, carefully review the results to ensure the desired anonymization is achieved.
3. If using generators and not transformers then using both JQ expression and default transformers have no side effects.
4. For critical fields that need specific anonymization, prefer using exact JQ paths over JQ expressions when possible.
