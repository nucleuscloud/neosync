<p align="center">
  <!-- <img alt="neosyncbanner" src="https://assets.nucleuscloud.com/neosync/docs/readme_header_new_logo.png"> -->
  <img alt="neosyncbanner" src="https://assets.nucleuscloud.com/neosync/docs/neosync-main-header-animated.svg" >
</p>

<p align="center" style="font-size: 24px;font-weight: 500;">
Open Source Synthetic Data Orchestration
<p>

<div align='center'>
 | <a href="https://neosync.dev">Website</a> |
 <a href="https://docs.neosync.dev">Docs</a> |
   <a href="https://discord.gg/HwrxVfNk">Discord</a> |
 <a href="https://neosync.dev/blog">Blog</a> |
 <a href="https://docs.neosync.dev/changelog">Changelog</a> |
 </div>

 <br>

<div align="center">
<p >
  <a href='http://makeapullrequest.com'>
    <img alt='PRs Welcome' src='https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=shields'/>
  </a>
  <img src="https://img.shields.io/github/license/lightdash/lightdash" />
  <!-- <a href="https://codecov.io/gh/nucleuscloud/neosync">
    <img alt="CodeCov" src="https://codecov.io/gh/nucleuscloud/neosync/graph/badge.svg?token=A35QDLRU04"/>
    </a> -->
     <a href="https://github.com/nucleuscloud/neosync/actions/workflows/main-tests.yml/">
    <img alt="main tests" src="https://github.com/nucleuscloud/neosync/actions/workflows/main-tests.yml/badge.svg"/>
    </a>
      <a href="https://x.com/neosynccloud">
    <img alt="Follow X" src="https://img.shields.io/twitter/follow/neosynccloud?label=Follow"/>
  </a>
</p>
</div>

## Introduction

[Neosync](https://neosync.dev) is a developer-first way to create anonymized or synthetic data and sync it across all environments.

Companies use Neosync to:

1. **Unblock local development** - Give developers the ability to self-serve de-identified and synthetic data whenever they need it
2. **Fix broken staging environments** - Catch bugs before they hit production when you hydrate your staging and QA environments with production-like data
3. **Keep environments in sync** - Keep your environments in sync with the latest synthetic data so you never hear "it works for me locally" again
4. **Get frictionless security, privacy and compliance** - Easily comply with laws like HIPAA, GDPR, and DPDP with de-identified and synthetic
5. **Seed development databases** - Easily seed development databases with synthetic data for unit testing, demos and more

## Features

- **Generate synthetic data** based on your schema
- **Anonymize existing production-data** to protect data
- **Subset your production database** for local and CI testing by filtering on an object, id or custom query
- **Complete async pipeline** that automatically handles job retries, failures and playback using an event-sourcing model
- **Referential integrity** for your data automatically - never worry about broken foreign keys again
- **Declarative, GitOps based configs** as a step in your CI pipeline to hydrate your CI DB
- **Pre-built data transformers** for all major data types
- **Custom data transformers**
- **Pre-built integrations** with Postgres, Mysql, S3

## Getting started

Neosync is a fully dockerized setup which makes it easy to get up and running.

We provide a `compose.yml` file that contains production image references that allow you to get up and running with just a few commands without having to build anything on your system.

To start Neosync, clone the repo into a local directory and then run:

```sh
make compose-up
```

To stop, run:

```sh
make compose-down
```

Neosync will now be available on [http://localhost:3000](http://localhost:3000).

## Kubernetes, Auth Mode and more

For more in-depth details on environment variables, Kubernetes deployments, and a production-ready guide, check out the [Deploy Neosync](https://docs.neosync.dev/deploy/introduction) section of our Docs.

## Resources

Some resources to help you along the way:

- [Docs](https://docs.neosync.dev) for comprehensive documentation and guides: Note these are still a work in progress.
- [Discord](https://discord.gg/HwrxVfNk) for discussion with the community and Neosync team
- [X](https://x.com/neosynccloud) for the latest updates

## Contributing

We love contributions big and small. Here are just a few ways that you can contribute to Neosync.

- Join our [Discord](https://discord.gg/HwrxVfNk) channel and ask us any questions there
- Open a PR (see our instructions on [developing with Neosync locally](https://docs.neosync.dev/guide/neosync-local-dev))
- Submit a [feature request](https://github.com/nucleuscloud/neosync/issues/new?assignees=&labels=enhancement%2C+feature&template=feature_request.md) or [bug report](https://github.com/nucleuscloud/neosync/issues/new?assignees=&labels=bug&template=bug_report.md)

## Mission

Our mission is to help developers build better, more resilient applications while protecting sensitive data. To do that, we built Neosync to give teams three things:

1. A world-class developer experience that fits into any workflow and follows modern developer best practices such as GitOps
2. A platform that can anonymize sensitive data or automatically generate synthetic data from a schema and sync that across all environments
3. An open source approach that allows you to keep your most sensitive data in your infrastructure

## Licensing

We strongly believe in free and open source software and make this repo is available under the [MIT expat license](./LICENSE.md).
