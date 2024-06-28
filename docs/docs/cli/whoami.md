---
title: whoami
description: Learn how to display the currently logged in user with the neosync whoami CLI command.
id: whoami
hide_title: true
slug: /cli/whoami
---

# neosync whoami

## Overview

Learn how to display the currently logged in user with the neosync whoami CLI command.

The `neosync whoami` command is used to show the currently logged in user.

## Usage

```bash
neosync whoami
```

## Options

The following options can be passed using the `neosync whoami` command:

- `--api-key` - Neosync API Key. Takes precedence over `$NEOSYNC_API_KEY`

## Environment Variables

| Variable        | Description                                                                                              | Is Required | Default Value         |
| --------------- | -------------------------------------------------------------------------------------------------------- | ----------- | --------------------- |
| NEOSYNC_API_URL | The base url of the Neosync API. This can be overridden to connect to different Neosync API environments | false       | http://localhost:8080 |
| NEOSYNC_API_KEY | The api key for Neosync API.                                                                             | false       |                       |
