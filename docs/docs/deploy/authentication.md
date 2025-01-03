---
title: Authentication
description: Learn how to configure and maintain authentication within Neosync open source
id: auth
hide_title: false
slug: /deploy/authentication
---

## Introduction

By default, Neosync launches without requiring any form of authentication. This is to make getting started with Neosync easier, and it's not advised to run Neosync in production without any form of authentication.

There are a few different authentication systems and play here, and this doc will detail what each one is, and why it exists.
There will also be a section for how to properly set up these systems properly.

## User Authentication

> **NB:** This requires a valid Neosync Enterprise license for OSS deployments. If you would like to try this out, please contact us.

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

## How to Authenticate against Neosync API

The API expects the standard `Authorization` header to come in via the HTTP request. The format should be `Bearer <token>`. This is true for both API Keys as well as user JWTs.

## APP Auth Configuration

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

## Auth Server Admin Access

The backend requires minimal admin access to the auth server in order to show information about members within an account.
If that is not needed or desired, the `AUTH_API_*` environment variables can be omitted, however the member page will not show any user data for team members.

Neosync currently supports keycloak and auth0 for this feature.

This is determined by the `AUTH_API_PROVIDER` environment variable that recognizes `auth0` and `keycloak` as their values. If omitted, `auth0` is the default for backwards compatibility.

The following environment variables are as follows:

- `AUTH_API_BASEURL` - This is the base url for the Admin API. Auth0 calls this the Management API, while Keycloak the Admin API.
  - For auth0, this is almost always your raw tenant url as custom domains do not work with Auth0 management API access. Example: `https://nucleus-cloud-staging.us.auth0.com`
  - For keycloak, this url will look something like this: `https://auth.svcs.stage.neosync.dev/admin/realms/neosync-stage`. The pattern is: `<baseurl>/admin/realms/<realm>`
- `AUTH_API_CLIENT_ID` - The service account's client id
- `AUTH_API_CLIENT_SECRET` - The client id secret

Scopes:

Today, this client only requires minimal access to the API to read users.
For Auth0, the service account should have the `read:users` scope under the `Auth0 Management API` audience.
For Keycloak, the `view-users` scope should be added to the service account roles, which can be found under the `realm-management` client scopes.

## Starting Neosync in Auth Mode

> **NB:** This requires a valid Neosync Enterprise license to be present in the API container. If you would like to try this out, please contact us.

Starting Neosync in Auth Mode is done in a similar way as starting Neosync in non-auth mode: using a compose file. A compose file is also provided that stands up [Keycloak](https://keycloak.org), an open source auth solution.

To stand up Neosync with auth, simply run the following command from the repo root:

```sh
make compose/auth/up
```

To stop, run:

```sh
make compose/auth/down
```

Neosync will now be available on [http://localhost:3000](http://localhost:3000) with authentication pre-configured!
Click the login with Keycloak button, register an account (locally) and you'll be logged in! -->
