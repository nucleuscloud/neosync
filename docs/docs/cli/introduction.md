---
title: Introduction
description: This section details Neosync CLI and all of its available commands.
id: intro
hide_title: false
slug: /cli/introduction
---

# CLI Overview

## Introduction

This section details Neosync CLI and all of its available commands.

```console
➜  ~ neosync
Terminal UI that interfaces with the Neosync system.

Usage:
  neosync [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  jobs        Parent command for jobs
  login       Login to Neosync
  sync        One off sync job to local resource
  version     Print the client version information
  whoami      Find out who you are

Flags:
      --api-key string   Neosync API Key. Takes precedence over $NEOSYNC_API_KEY
      --config string    config file (default is $HOME/.neosync/neosync.yaml)
  -h, --help             help for neosync
  -v, --version          version for neosync

Use "neosync [command] --help" for more information about a command.
```

## Environment Variables

There are a few global environment variables that are available on every request.

### `NEOSYNC_API_KEY`

Neosync API Key. Used if logging in via a system api key.

### `NEOSYNC_API_URL`

The url of the Neosync API to direct the request to.

### Persisting CLI Environment Variables

Environment Variables for the CLI may be persisted by setting them in the config file.
By default this is located at `$HOME/.neosync/config.yaml`.
The CLI does respect `XDG_CONFIG_HOME` as well as a `NEOSYNC_CONFIG_DIR` may be optionally set to override the default location.

Example of a config.yaml:

```yaml
NEOSYNC_API_URL: 'http://localhost:8080'
```

The CLI uses [viper](https://github.com/spf13/viper) for environment management, and has various configuration options that come with it.
You can find the environment setup method [here](https://github.com/nucleuscloud/neosync/blob/main/cli/internal/cmds/neosync/neosync.go#L80).

### Full list

For a full list of environment variables and flags available, see the specific command you are running.
Otherwise, there is a top-level list of all environment variables spread across all commands available [here](../deploy/environment-variables.md#cli).

## Metadata

CLI metadata is appended to the outgoing gRPC context and HTTP Headers to provide tracking and metadata to the API.
This lets the API know which version the CLI is using when it invokes commands to better track CLI usage over time.

The following metadata is added to all CLI context:

- Git Version
- Git Commit
- OS Platform
