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
  <img alt="Follow twitter" src="https://img.shields.io/twitter/follow/neosynccloud?label=Follow"/>
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
- [Licensing](#licensing)

## Getting started

You can also check out our [Docs](https://docs.neosync.dev) for more guides including a production-ready guide.

## Run Neosync locally

To set up and run Neosync locally, make sure you have Git and Docker installed on your system.


### Tools
Currently, the primary development environment is done by deploying the app and its dependent resources into a `kind` cluster using `tilt`.
We utilize `helm` charts to wrap up deployable artifacts and have `tilt` install these to closely mimic a production environment.
Due to this, there are a number of dependencies that must be installed on a host system prior to being able to run `neosync` locally.
We are working on making this smoother by providing a `devcontainer` that comes pre-installed with all of the dependencies required.

Detailed below are the main dependencies are descriptions of how they are utilized:

#### Kubernetes
Kubernetes is used today as our primary development environment. Tilt is a great tool that lets you define your environment in code.
This lets us develop quickly, locally, while closely mimicking a real production environment.
* [kind](https://github.com/kubernetes-sigs/kind)
  * Kubernetes in Docker. We use this to spin up a barebones kubernetes cluster that deploys all of the `neosync` resources.
* [tilt](https://github.com/tilt-dev/tilt)
  * Allows us to define our development environment as code.
* [ctlptl](https://github.com/tilt-dev/ctlptl)
  * CLI provided by the Tilt-team to make it easy to declaratively define the kind cluster that is used for development
* [kubectl](https://github.com/kubernetes/kubectl)
  * Allows for observability and management into the spun-up kind cluster.
* [kustomize](https://github.com/kubernetes-sigs/kustomize)
  * yaml templating tool for ad-hoc patches to kubernetes configurations
* [helm](https://github.com/helm/helm)
  * Kubernetes package manager. All of our app deployables come with a helm-chart for easy installation into kubernetes
* [helmfile](https://github.com/helmfile/helmfile)
  * Declaratively define a helmfile in code! We have all of our dev charts defined as a helmfile, of which Tilt points directly to.

#### Golang + Protobuf
* Golang
  * The language of choice for our backend and worker packages
* [sqlc](https://github.com/sqlc-dev/sqlc)
  * Our tool of choice for the data-layer. This lets us write pure SQL and let sqlc generate the rest.
* [buf](https://github.com/bufbuild/buf)
  * Our tool of choice for interfacing with protobuf
* [golangci-ci](https://github.com/golangci/golangci-lint)
  * The golang linter of choice

#### Npm/Nodejs
* Node/Npm


All of these tools can be easily installed with `brew` if on a Mac.
Today, `sqlc` and `buf` don't need to be installed locally as we exec docker images for running them.
This lets us declare the versions in code and docker takes care of the rest.

It's of course possible run everything on bare metal without Kuberentes or Tilt, but there will be more work getting everything up and running (at least today).

### Brew Install
Each tool above can be straightforwardly installed with brew if on Linux/MacOS
```
brew install kind tilt-dev/tap/tilt ctlptl kubernetes-cli kustomize helm helmfile go sqlc buf golangci-lint node
```

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
We've spent time making the development process smooth with kind and Tilt.
However, we understand not everyone wants to develop or is comfortable working inside of a Kubernetes cluster.

We'd like to support a pure docker compose flow, but haven't had the time to invest into providing that experience yet.
If this is a killer development feature for you, let us know by submitting a feature request!

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
- Submit a [feature request](https://github.com/nucleuscloud/neosync/issues/new?assignees=&labels=enhancement%2C+feature&template=feature_request.md) or [bug report](https://github.com/nucleuscloud/neosync/issues/new?assignees=&labels=bug&template=bug_report.md)

## Licensing

We strongly believe in free and open source software and make this repo is available under the [MIT expat license](./LICENSE.md).
