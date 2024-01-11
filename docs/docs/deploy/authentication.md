---
title: Authentication
id: auth
hide_title: false
slug: /deploy/authentication
---

## Introduction

Authentication is a core pillar to a secure, productionized deployment of Neosync.
By default, Neosync launches without requiring any form of authentication. This is to make getting started with Neosync easier, and it's not advised to run Neosync in production without any form of authentication.

There are a few different authentication systems and play here, and this doc will detail what each one is, and why it exists.
There will also be a section for how to properly set up these systems properly.

## User Authentication

This authentication is the primary form of authentication in the system. It is used to identify authenticating users in the system and what privileges they have.
Today, this is quite simple and merely verifies their access token is valid and that they are operating against a Neosync account that a user resides in.

In the future, it is planned to add more fine-grained authentication in the form of scopes. This will enable more granular authorization against specific actions in the system against targeted resources.

This form of authentication is configured with any Open Identity Connect (OIDC) protocol compliant provider.
Internally, we utilize Auth0, and to date this is the only provider that has been fully tested for use with Neosync.

Environment variables must be provided for the App ad API to properly configure User Authentication.
See the [environment variables](/deploy/environment-variables.md) page for the whole list of required auth env vars.
The table shows that most of them are not required, however that is only true if `AUTH_ENABLED` is set to false. They must be provided if `AUTH_ENABLED=true`.

The descriptions of each environment variable detail exactly what the `AUTH_*` is and why it is needed.

If there is any trouble configuring authentication with Neosync, reach out to us for help on Discord.

### API Auth Configuration

The backend requires less configuration as it simply needs to validate the incoming JWTs. However, there are a few Auth0-specific environment variables that are used if the authenticated user does use Auth0.
The backend will look up more information for them via the Auth0 management APIs to show things like user photo, email, etc. on the Members page in the App. As more auth providers are added, this will expand and change based on which provider is configured.

#### Auth0 Management Client

This section details the `AUTH_API_*` environment variables.

These three environment variables: `AUTH_API_BASEURL`, `AUTH_API_CLIENT_ID`, `AUTH_API_CLIENT_SECRET` are used by the API to communicate with Auth0 to retrieve user information that is surfaced on the Members page.
In Auth0's case, this client should be a service account.

If using Auth0, the required scope for the `Auth0 Management API` is `read:users`.

#### How to Authenticate against the API

The API expects the standard `Authorization` header to come in via the HTTP request. The format should be `Bearer <token>`. This is true for both API Keys as well as user JWTs.

### APP Auth Configuration

The app requires a bit more configuration since there is also session management that occurs with the help of `next-auth`.
Today, the App has only been tested with Auth0, however, theoretically any auth provider credentials could be entered for the environment variable values and it should work.
We are working towards making this more generic to allow more providers like Keycloak, Google, etc.

## API Key Authentication

This authentication is the primary form of authentication for system-level access. It is used to identify and authenticate machine users.
Today, the primary function for this is for use with the Worker process, as well as CLI-access in an automated environment like Github Actions.
This form of authentication also makes writing scripts with the Neosync SDKs much easier and it isn't as easy to provide a user access token.

API Keys have their own Neosync User Identifier assigned to them and are scoped to the Neosync Account.
They have a maximum expiration of 1 year before they expire and require rotation.
It's advised to rotate the keys prior to that, or simply create a new one to allow overlap as once the key has been rotated, the old one will no longer work.

### Configuration

It's important to note here that API Key is not enabled unless the `AUTH_ENABLED` environment variable is set to `true` in the API.
To enable auth, User authentication must be properly configured as well. Both systems are turned on or off.

An API Key can be created either through the SDK or via the web app.

To do so via the web app:

1. navigate to the relevant instance of Neosync
2. go to the `Settings` page for the account that it is desired to create an API key for.
3. Click the API Keys section
4. Click the `+ New API Key` button.
5. Write down a name and select when it should expire
6. Submit

If successful, you should now be on the API Key Details page and the API Key should be seen in plaintext on the page.

It's important to save this somewhere as it is no longer retrievable again. If lost, a new key must be regenerated.
These keys are not stored in plaintext in the database and are one-way hashed so the original contents are no longer retrievable.

## Temporal mTLS Authentication

Neosync API and Neosync Worker both require mTLS authentication when interfacing with Temporal (if this is enabled in Temporal).
Like Neosync, by default, the local versions of Temporal don't require authentication by default (although this may be changing.)
If using Temporal Cloud, mTLS is required by default and must be configured to properly communicate with the Temporal servers.

### mTLS Certificate Configuration

Temporal has a guide for creating mTLS certs [here](https://docs.temporal.io/cloud/certificates#use-tcld-to-generate-certificates).

Once these have been created, they must be provided as environment variables to both the API and Worker processes.
Reference the [environment variables](/deploy/environment-variables.md) page for the `TEMPORAL_*` environment variables.
