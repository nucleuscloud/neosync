---
title: Using Neosync in CI
description: Learn how to use Neosync in your CI pipelines in order to hydrate your CI databases with anonymized production data
id: using-neosync-in-ci
hide_title: false
slug: /guides/using-neosync-in-ci
---

## Introduction

Continuous Integration is a primary usecase for utilizing Neosync. It's often the case that integration or unit tests run in CI that need good data.
It's easy enough to spin up a Postgres or other kind of database using Github Actions, but the problem becomes hydrating that database with solid data that can be used for testing purposes.

For this reason, we built the [neosync sync](../cli/sync.md) command to enable synchronizing a connection configured in Neosync to a locally hosted database, or any other database that may not otherwise be available over the internet easily.

Getting the CLI installed into a Github Action is task in itself, however. This is why we are offering a first class way to install the Neosync CLI.

## Neosync CLI Github Action

We've built the [Setup Neosync CLI Action](https://github.com/nucleuscloud/setup-neosync-cli-action) Github Action that can be easily dropped into any job to immediately get setup with the Neosync CLI.
The README for that action gives detailed instructions on how to get that set up. Afterwards, any `neosync` command can be run in subsequent jobs.

## Setup a Github Action to sync remote data to a CI Postgres Database

This is a full example of a Github Action that pulls down data from a remotely configured Neosync Connection and hydrates the local Postgres database.

If you want to try this with a sample dataset, there is a good one on [sqltutorial.org](https://www.sqltutorial.org/sql-sample-database/)

```yaml
name: Setup Neosync CLI and PostgreSQL Database
​
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
​
jobs:
  setup:
    runs-on: ubuntu-latest
​
    services:
      postgres:
        image: postgres
        env:
          POSTGRES_DB: neosync
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
        # Set health checks to wait until postgres has started
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
​
    steps:
      - name: Checkout
        uses: actions/checkout@v4
​
      - name: PostgreSQL Schema Setup
        run: |
          PGPASSWORD=postgres psql -h localhost -U postgres -d neosync -f sql/create.sql

      - name: Select from Employees table to see that it's empty
        run: |
          PGPASSWORD=postgres psql -h localhost -U postgres -d neosync -c 'SELECT * from neosync.employees;'
​
      - name: Set up Neosync CLI
        uses: nucleuscloud/setup-neosync-cli-action@v1
​
      - name: Neosync sync command
        run: neosync sync --api-key ${{ secrets.NEOSYNC_API_KEY }} --connection-id <connection-uuid> --destination-driver postgres --destination-connection-url "postgresql://postgres:postgres@localhost:5432/neosync?sslmode=disable"
        env:
          NEOSYNC_API_URL: <neosync-api-url>
​
      - name: Select from PostgreSQL
        run: |
          PGPASSWORD=postgres psql -h localhost -U postgres -d neosync -c 'SELECT * from neosync.employees;'
```
