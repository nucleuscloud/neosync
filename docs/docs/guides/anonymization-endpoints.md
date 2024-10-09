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

## AnonymizeSingle, AnonymizeMany

### Transformer Mappings
Transformer mappings define specific paths in your JSON data using JQ and the transformers to apply to those paths.

### Default Mappings
Default transformers provide a way to anonymize fields that are not explicitly defined in the transformer mappings. They work on a data type basis, applying to all fields of a specific type that don't have a transformer mapping configured. It supports default transformers for three data types: strings, numbers, and booleans.

Specific transformer mappings always take precedence over default transformers. If a field has a transformer defined in the mappings, the default transformer for that data type will not be applied to that field.

This approach allows 

> If using a JQ expression (.. | objects | select(has("name")) | .name) and not a path (.path.to.field) then the default transformers will override the values that the expression finds.
