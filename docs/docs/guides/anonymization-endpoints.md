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
Default transformers apply to all data types not explicitly defined in the transformer mappings. Transformers are mapped by data type. 

