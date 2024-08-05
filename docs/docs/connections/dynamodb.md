---
title: DynamoDB
description: Amazon DynamoDB is a fully managed proprietary NoSQL database offered by Amazon.com as part of the Amazon Web Services portfolio.
id: dynamodb
hide_title: false
slug: /connections/dynamodb
# cSpell:words textareas
---

## Introduction

DynamoDB offers a fast persistent key–value datastore with built-in support for replication, autoscaling, encryption at rest, and on-demand backup among other features. It is one of the most highly requested database connections to be added to Neosync.

If you are interested in using DynamoDB but don't see a feature that is required for you to use it, please reach out to us on Discord!

## Configuring DynamoDB

<!-- todo: talk about the IAM permissions that are required, and the different ways to authenticate to DynamoDB -->

## Permissions

Different permissions may be utilized when configuring your DynamoDB instance as a source or destination connection as they require different things.

## Source Connections

Ensure that the role or access credentials that are used here have read permissions to the DynamoDB tables that will be read from.

## Destination Connections

Ensure that the role or access credentials that are used here have read and write permissions to the DynamoDB tables that will be written to.
