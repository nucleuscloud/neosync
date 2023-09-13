<p align="center">
  <img alt="neosyncbanner" src="https://assets.nucleuscloud.com/neosync/neosyncreadmelight.png">
    <!-- <img alt="neosyncbanner" src="https://assets.nucleuscloud.com/neosync/neosync_readme_banner.svg"> -->
</p>

<p align="center" style="font-size: 24px">
Open source Test Data Management
<p>

<p align="center" style="font-size: 14px">
Neosync is a developer-first way to create anonymized, secure test data and sync it across all environments for high-quality local, stage and CI testing
<p>

<p align="center">
  <a href='http://makeapullrequest.com'><img alt='PRs Welcome' src='https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=shields'/></a>
<img src="https://img.shields.io/github/license/lightdash/lightdash" />
  <img alt="Follow twitter" src="https://img.shields.io/twitter/follow/neosynccloud?label=Follow
"/>
</p>

<p align="center">
  <a href="https://docs.neosync.dev">Docs</a> - <a href="https://neosync.dev/slack">Community</a> - <a href="https://neosync.dev/roadmap">Roadmap</a> - <a href="https://neosync.dev/changelog">Changelog</a> 
</p>

## Introduction

[Neosync](https://neosync.dev) is an open source platform that connects to a snapshot of your proudction database and allows teams to either generate synthetic data from that production schema or anonymize production-data . generate synthetic data from their produteams use to generate anonymized, production-like data and sync it across all of their environments for high-quality local, stage and CI testing.

Our mission is to help developers build better, more resilient applications while protecting sensitive data. To do that, we built Neosync to give teams three things:

1. A world-class developer experience that fits into any workflow and follows modern developer best practices such as GitOps
2. A platform that can anonymize sensitive data or automatically generate synthetic data from a schema and sync that across all environments
3. An open source approach that allows you to keep your most sensitive data in your infrastructure

## Features

- Automatically generate synthetic data based on your schema
- Anonymize existing production-data to protect data
- Create subsets of your production database for local and CI testing by filtering on an object, id or custom query
- Complete async pipeline that automatically handles job retries, failures and playback using an event-sourcing model
- Complete referential integrity for your data automatically - never worry about broken foreign keys again
- APIs and SDKs so you can build your own workflows to hydrate non-prod DBs
- Use our declarative, GitOps based configs as a step in your CI pipeline to hydrate your CI DB
- Pre-built transformers for all major data types
- Define custom transformers in code to transform your data in any way you want
- Pre-built integrations with Postgres, Mysql, Mongo, S3, Big Query, Snowflake and much more

## Table of Contents

- [Getting Started](#get-started-for-free)
- [Features](#features)
- [Resources](#docs-and-support)
- [Contributing](#contributing)
- [Licenseing](#licensing)

## Getting started

You can also check out our [Docs](https://docs.neosync.dev) for more guides including a production-ready guide.

### Run Neosync locally

To set up and run Neosync locally, make sure you have Git and Docker installed on your system. Then run the following command based on what type of machine you have.

Linux/macOS

```bash
git clone https://github.com/nucleuscloud/neosync && cd "$(basename $_ .git)" && cp .env.example .env && docker-compose -f docker-compose.yml up
```

## Resources

Some resources to help you along the way:

- [Docs](https://docs.neosync.dev) for comprehensive documentation and guides
- [Slack](https://neosync.dev) for discussion with the community and Neosync team
- [Github](https://github.com/nucleuscloud/neosync)
- [Twitter](https://twitter.com/neosyncloud) for the latest updates
- [Blog](https://neosync.com/blog) for insights, articles and more
- [Roadmap](https://neosync.dev/roadmap) for future development

## Contributing

We love contributions big and small. Here are just a few ways that you can contirbute to Neosync.

- Vote on features or get early access to beta functionality in our [roadmap](https://neosync.dev/roadmap)
- Join our [Slack](https://neosync.dev) channel and ask us any questions there
- Open a PR (see our instructions on [developing with Neosync locally](https://docs.neosync.dev/developing-locally))
- Submit a [feature request](https://github.com/PostHog/posthog/issues/new?assignees=&labels=enhancement%2C+feature&template=feature_request.md) or [bug report](https://github.com/PostHog/posthog/issues/new?assignees=&labels=bug&template=bug_report.md)

We publish all contributions in a monthly newsletter and send that out to thousands of people to show our gratitude for our community.

## Licensing

We strongly believe in free and open source software and make this repo is available under the [MIT expat license](https://github.com/nucleuscloud/neosync/blob/main/LICENSE).
