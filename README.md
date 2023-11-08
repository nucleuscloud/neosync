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
  <a href='http://makeapullrequest.com'>
    <img alt='PRs Welcome' src='https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=shields'/>
  </a>
  <img src="https://img.shields.io/github/license/lightdash/lightdash" />
  <a href="https://x.com/neosynccloud">
    <img alt="Follow X" src="https://img.shields.io/twitter/follow/neosynccloud?label=Follow"/>
  </a>
  <a href="https://codecov.io/gh/nucleuscloud/neosync">
    <img alt="CodeCov" src="https://codecov.io/gh/nucleuscloud/neosync/graph/badge.svg?token=A35QDLRU04"/>
  </a>
</p>

<!-- <p align="center">
  <a href="https://docs.neosync.dev">Docs</a> - <a href="https://neosync.dev/slack">Community</a> - <a href="https://neosync.dev/roadmap">Roadmap</a> - <a href="https://neosync.dev/changelog">Changelog</a>
</p> -->

<strong>Heads up! This repo is still in development mode while we continue to build out the first stable version. In the mean-time, we'll happily accept PRs and support! </strong>

## Introduction

![neosync-data-flow](https://assets.nucleuscloud.com/neosync/readmeheader.svg)

[Neosync](https://neosync.dev) is an open source platform that connects to a snapshot of your production database and allows teams to either generate synthetic data from their production schema or anonymize production-data and sync it across all of their environments for high-quality local, stage and CI testing.

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
- [Licensing](#licensing)

## Getting started

You can also check out our [Docs](https://docs.neosync.dev) for more guides including a production-ready guide. Note: these are still a work in progress.

## Run Neosync locally

To set up and run Neosync locally, make sure you have Git and Docker installed on your system.

The sections below detail the tools required to build and run the Neosync development environment.
This can be circumvented by using our official devcontainer which comes pre-installed with all of the tools necessary.

There are then two ways to start Neosync:

- Tilt
- Docker Compose

Tilt is the method we currently use to development Neosync. This lets us develop as if we are running inside of a Kubernetes cluster.
This isn't for everyone, which is we also offer a compose method for a simpler, kubernetesless approach.

Check out the sections below for which method applies to you.

### Tools

Currently, the primary development environment is done by deploying the app and its dependent resources into a `kind` cluster using `tilt`.
We utilize `helm` charts to wrap up deployable artifacts and have `tilt` install these to closely mimic a production environment.
Due to this, there are a number of dependencies that must be installed on a host system prior to being able to run `neosync` locally.

Detailed below are the main dependencies are descriptions of how they are utilized:

#### Kubernetes

Kubernetes is used today as our primary development environment. Tilt is a great tool that lets you define your environment in code.
This lets us develop quickly, locally, while closely mimicking a real production environment.

- [kind](https://github.com/kubernetes-sigs/kind)
  - Kubernetes in Docker. We use this to spin up a barebones kubernetes cluster that deploys all of the `neosync` resources.
- [tilt](https://github.com/tilt-dev/tilt)
  - Allows us to define our development environment as code.
- [ctlptl](https://github.com/tilt-dev/ctlptl)
  - CLI provided by the Tilt-team to make it easy to declaratively define the kind cluster that is used for development
- [kubectl](https://github.com/kubernetes/kubectl)
  - Allows for observability and management into the spun-up kind cluster.
- [kustomize](https://github.com/kubernetes-sigs/kustomize)
  - yaml templating tool for ad-hoc patches to kubernetes configurations
- [helm](https://github.com/helm/helm)
  - Kubernetes package manager. All of our app deployables come with a helm-chart for easy installation into kubernetes
- [helmfile](https://github.com/helmfile/helmfile)
  - Declaratively define a helmfile in code! We have all of our dev charts defined as a helmfile, of which Tilt points directly to.

#### Golang + Protobuf

- Golang
  - The language of choice for our backend and worker packages
- [sqlc](https://github.com/sqlc-dev/sqlc)
  - Our tool of choice for the data-layer. This lets us write pure SQL and let sqlc generate the rest.
- [buf](https://github.com/bufbuild/buf)
  - Our tool of choice for interfacing with protobuf
- [golangci-ci](https://github.com/golangci/golangci-lint)
  - The golang linter of choice

#### Npm/Nodejs

- Node/Npm

All of these tools can be easily installed with `brew` if on a Mac.
Today, `sqlc` and `buf` don't need to be installed locally as we exec docker images for running them.
This lets us declare the versions in code and docker takes care of the rest.

It's of course possible run everything on bare metal without Kuberentes or Tilt, but there will be more work getting everything up and running (at least today).

### Brew Install

Each tool above can be straightforwardly installed with brew if on Linux/MacOS

```
brew install kind tilt-dev/tap/tilt tilt-dev/tap/ctlptl kubernetes-cli kustomize helm helmfile go sqlc buf golangci-lint node
```

### Devcontainer

Host machine setup can be skipped by developing inside of a vscode devcontainer.
This container comes pre-baked with all of the tools we use to develop and work on neosync.
This container also supports running neosync with compose or tilt.

### Setup with Tilt

Step 1 is to ensure that the `kind` cluster is up and running along with its registry.
This can be manually created, or done simply with `ctlptl`.
The cluster is declaratively defined [here](./tilt/kind/cluster.yaml)

Note: Because databases are installed into the Kubernetes cluster, we like to persist them directly on the host volume to survive cluster re-creations.
We mount a container path locally in a `.data` folder. If on a Mac, ensure that you've allowed wherever this repository has been cloned into the allow-list in Docker Desktop.

The below command invokes the cluster-create script that can be found [here](./tilt/scripts/cluster-create.sh)

```
make cluster-create
```

After the cluster has been successfully created, `tilt up` can be run to start up `neosync`.
Refer to the top-level [Tiltfile](./Tiltfile) for a clear picture of everything that runs.
Each dependency in the `neosync` repo is split into sub-Tiltfiles so that they can be run in isolation, or in combination with other sub-resources more easily.

Once everything is up and running, the app can be accessed at locally at `http://localhost:3000`.

### Setup with Docker Compose

Neosync can be run with compose. This works pretty well, but is a bit more manual today than with Tilt.
Not everything is hot-reload, but you can successfully run everything using just compose instead of having to manage a kubernetes cluster and running Tilt.
To enable hot reloading, must run `docker compose watch` instead of `up`. Currently there is a limitation with devcontainers where this command must be run via `sudo`.

There are two compose files that need to be run today. The first is the Temporal compose, the second is the Neosync compose.
It's suggested you run these separate (as of today) for a clean separation of concerns.

```
$ docker compose -f temporal/compose.yml up -d
$ docker compose -f compose.yml up -d
```

Once everything is up and running, the app can be accessed locally at `http://localhost:3000`.

Work to be done:

- inherit the temporal compose inside of the neosync compose, separate with compose profiles.

## Resources

Some resources to help you along the way:

- [Docs](https://docs.neosync.dev) for comprehensive documentation and guides: Note these are still a work in progress.
- [Slack](https://join.slack.com/t/neosynccommunity/shared_invite/zt-1o3g7cned-OTBKzNj3InDm1YmpYdqRXg) for discussion with the community and Neosync team
- [X](https://x.com/neosynccloud) for the latest updates
<!-- - [Blog](https://neosync.com/blog) for insights, articles and more
- [Roadmap](https://neosync.dev/roadmap) for future development -->

## Contributing

We love contributions big and small. Here are just a few ways that you can contribute to Neosync.

<!-- - Vote on features or get early access to beta functionality in our [roadmap](https://neosync.dev/roadmap) -->

- Join our [Slack](https://join.slack.com/t/neosynccommunity/shared_invite/zt-1o3g7cned-OTBKzNj3InDm1YmpYdqRXg) channel and ask us any questions there
- Open a PR (see our instructions on [developing with Neosync locally](https://docs.neosync.dev/developing-locally))
- Submit a [feature request](https://github.com/nucleuscloud/neosync/issues/new?assignees=&labels=enhancement%2C+feature&template=feature_request.md) or [bug report](https://github.com/nucleuscloud/neosync/issues/new?assignees=&labels=bug&template=bug_report.md)

## Licensing

We strongly believe in free and open source software and make this repo is available under the [MIT expat license](./LICENSE.md).

## Triggering a Release

Triggering a release is done by cutting a git tag.
This causes all artifacts for the various components in the system to build and publish.

### Tag Format:

The tag format is a semver compliant tag that starts with `v`.

Examples:

- `v0.0.1`
- `v0.0.1-nick.1`

This is done by running the `hack/tag.sh` script like so:

```sh
$ ./hack/tag.sh <tag>
```

Example:

```sh
$ ./hack/tag.sh v0.0.1
```
