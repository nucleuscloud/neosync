---
title: Configuring Neosync with Terraform
id: terraform
hide_title: false
slug: /guides/terraform
---

## Introduction

Neosync ships with an official Terraform provider that can be used to create, read, update, and delete supported Neosync resources.

## Setup and Configuration

Before configuring the Neosync provider, you'll need to know a few pieces of config data so that you can properly configure the provider.

### Endpoint Url

#### _Neosync Cloud_

If configuring via Neosync Cloud, no endpoint is required as the provider already defaults to the Neosync Cloud instance.

#### _Self-Hosted_

If self-hosting, the url to your self-hosted instance of Neosync API must be provided.
It can be provided either directly to the provider as a configuration parameter, or via the `NEOSYNC_ENDPOINT` environment variable. This is detailed in the Terraform Registry docs as well.

### API Key

Next, you'll need to generate an API key that the Terraform provider can use to act on the behalf of your account.
If you haven't configured one, you can do so by heading over to the api key page in the settings for your specific account and creating one.

If the self-hosted instance is running without authentication enabled, this API Key is not utilized, but an account-id must be provided.

The API Key may be input as a variable to the provider, or provided in the environment through the `NEOSYNC_API_TOKEN` environment variable.

### Account Id

Generally, this option is ommitted as it is inferred through the API Key.
If self-hosting Neosync and running without authentication, or simply wanting to be redundant, provide the account id to the provider or via the `NEOSYNC_ACCOUNT_ID` environment variable to explicitly tell the provider which account id to use.

## Terraform Registry Docs

To dive deeper into the provider documentation, head over to the latest docs page on the [Terraform Registry](https://registry.terraform.io/providers/nucleuscloud/neosync/latest/docs).
The registry documentation provided a better look into what resources are available, as well as examples of how to use them.

## Bugs or Features

If there is an issue with the provider, or there is a feature that is missing, please do not hesitate to log an issue or join our Discord and ask about it.
