<p align="center">
  <img alt="neosyncbanner" src="https://assets.nucleuscloud.com/neosync/docs/readme_header_new_logo.png">
</p>

<p align="center" style="font-size: 24px;font-weight: 500;">
Open Source Synthetic Data Orchestration
<p>

<div align='center'>
 | <a href="https://neosync.dev">Website</a> |
 <a href="https://docs.neosync.dev">Docs</a> |
 <a href="https://neosync.dev/blog">Blog</a> |
 <a href="https://docs.neosync.dev/changelog">Changelog</a> |
 </div>

<p align="center" style='padding-top: 20px'>
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
 <a href="https://github.com/orgs/nucleuscloud/projects/6">
        <img alt="Roadmap" src="https://img.shields.io/badge/project%20roadmap-73C649"/>
 </a>
  </a>
</p>

<!-- <p align="center">
  <a href="https://docs.neosync.dev">Docs</a> - <a href="https://neosync.dev/slack">Community</a> - <a href="https://neosync.dev/roadmap">Roadmap</a> - <a href="https://neosync.dev/changelog">Changelog</a>
</p> -->

## Introduction

![neosync-data-flow](https://assets.nucleuscloud.com/neosync/docs/readmeheader.svg)

[Neosync](https://neosync.dev) is a developer-first way to create anonymized or synthetic data and sync it across all environments for high-quality local, stage and CI testing

Our mission is to help developers build better, more resilient applications while protecting sensitive data. To do that, we built Neosync to give teams three things:

1. A world-class developer experience that fits into any workflow and follows modern developer best practices such as GitOps
2. A platform that can anonymize sensitive data or automatically generate synthetic data from a schema and sync that across all environments
3. An open source approach that allows you to keep your most sensitive data in your infrastructure

## Features

- Automatically generate synthetic data based on your schema
- Anonymize existing production-data to protect data
- Create subsets of your production database for local and CI testing by filtering on an object, id or custom query
- Complete async pipeline that automatically handles job retries, failures and playback using an event-sourcing model
- Referential integrity for your data automatically - never worry about broken foreign keys again
- Use our declarative, GitOps based configs as a step in your CI pipeline to hydrate your CI DB
- Pre-built transformers for all major data types
- Define custom transformers
- Pre-built integrations with Postgres, Mysql, S3

## Getting started

You can also check out our [Docs](https://docs.neosync.dev) for more guides including a production-ready guide.

## Run Neosync locally

Neosync is a fully dockerized setup. Due to this, there are many different ways to run Neosync.

There are three officially supported ways of running Neosync locally:

1. Bare Metal
2. `docker compose`
3. Kubernetes via Tilt and Kind.

For more in-depth details on environment variables as well as Kubernetes deployments, check out the [Deploy Neosync](https://docs.neosync.dev/deploy/introduction) section of our Docs.

This readme will focus more on the development environment and simple steps to getting Neosync up on your system.

### Simply trying Neosync

If you just want to try out Neosync to see what it's like or get a feel for the product, most of the development setup guide below can be skipped.
We provide a `compose.yml` file that contains production image references that allow you to get up and running with just a few commands without having to build anything on your system.

The simplest configuration of Neosync is standing it up without any form of authentication.
This can be done with the following command:

```sh
make compose-up
```

To stop, run:

```sh
make compose-down
```

Neosync will now be available on [http://localhost:3000](http://localhost:3000).

A compose file is also provided that stands up [Keycloak](https://keycloak.org), an open source auth solution.

To stand up Neosync with auth, simply run the following command:

```sh
make compose-auth-up
```

To stop, run:

```sh
make compose-auth-down
```

Neosync will now be available on [http://localhost:3000](http://localhost:3000) with authentication pre-configured!
Click the login with Keycloak button, register an account (locally) and you'll be logged in!

### Neosync Development Environment

This section goes into detail each tool that is used for development Neosync.
This section casts a wide net, and some tools may not be required depending on if you are using a Tilt setup or a Compose setup.
Most of the `Kubernetes` focused tools can be skipped if develoing via compose.

### Neosync DevContainer

Neosync has a pre-published [devcontainer](https://containers.dev/) that can be used to easily get a working Neosync dev environment.
This container comes pre-packaged with all of the tools needed for developing Neosync, and works with Tilt or Compose, or Bare Metal setups.

### Tools

This section contains a flat list of the tools that are used to develop Neosync and why.

Detailed below are the main dependencies are descriptions of how they are utilized:

#### Kubernetes

If you're choosing to develop in a Tilt environment, this section is more important as it contains all of the K8s focused tooling.

Tilt is a great tool that is used to automate the setup of a Kubernetes cluster. There are multiple `Tiltfile`'s througout the code, along with a top-level one that is used to inject all of the K8s manifests to setup Neosync inside of a K8s cluster.

This enables fast development, locally, while closely mimicking a real production environment.

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

#### Go + Protobuf

- [Go](https://go.dev/)
  - The language of choice for our backend and worker packages
- [sqlc](https://github.com/sqlc-dev/sqlc)
  - Our tool of choice for the data-layer. This lets us write pure SQL and let sqlc generate the rest.
- [buf](https://github.com/bufbuild/buf)
  - Our tool of choice for interfacing with protobuf
- [golangci-ci](https://github.com/golangci/golangci-lint)
  - The golang linter of choice
- [migrate](https://github.com/golang-migrate/migrate)
  - Golang Migrate is the tool that is used to run DB Migrations for the API.

#### Npm/Nodejs

- [Node/Npm](https://nodejs.org/en)
  - Used to run the app, along with Nextjs.

All of these tools can be easily installed with `brew` if on a Mac.
Today, `sqlc` and `buf` don't need to be installed locally as we exec docker images for running them.
This lets us declare the versions in code and docker takes care of the rest.

### Brew Install

Each tool above can be straightforwardly installed with brew if on Linux/MacOS

```
brew install kind tilt-dev/tap/tilt tilt-dev/tap/ctlptl kubernetes-cli kustomize helm helmfile go sqlc buf golangci-lint node
```

### Setup with Compose

When running with either `Tilt` or `docker compose`, volumes are mapped from these filesystems to the host machine for both neosync and Temporal's databases.
A volume is mounted locally in a `.data` folder.

To enable hot reloading, must run `docker compose watch` instead of `up`. **Currently there is a limitation with devcontainers where this command must be run via `sudo`.**
This works pretty well with the `app`, but can be a bit buggy with the `api` or `worker`.
Sometimes it's a little easier to just rebuild the docker container like.

Assuming the latest binary is available in the bin folder:

```
$ docker compose up -d --build api
```

#### Building the backend and worker when using Docker Compose.

If using the dev-focused compose instead of the `*-prod.yml` compose files, the binaries for the `api` and `worker` will need to be built.

Run the following command to build the binaries:

```sh
make build
```

To Rebuild the binaries, run:

```sh
make rebuild
```

When building the Go processes with the intention to run with `docker compose`, it's important to run `make dbuild` instead of the typical `make build` so that the correct `GOOS` is specified. This is only needed if your native OS is not Linux (or aren't running in a devcontainer).
The `make dbuild` command ensures that the Go binary is compiled for Linux instead of the host os.

This will need to be done for both the `worker` and `api` processes prior to running compose up. Using the following command will build both the binaries for you:

```sh
make compose-dev-up
```

To stop, run:

```sh
make compose-dev-down
```

Once everything is up and running, the app can be accessed locally at [http://localhost:3000](http://localhost:3000).

#### Running Compose with Authentication

Note, a compose file with authentication pre-configured can be found [here](./compose/compose-auth.yml).
This will stand up Keycloak with a pre-configured realm that will allow logging in to Neosync with a standard username and password, completely offline!

```sh
make compose-dev-auth-up
```

To stop, run:

```sh
make compose-dev-auth-down
```

#### Docker Desktop

If using Docker Desktop, the host file path to the `.data` folder will need to be added to the File Sharing tab.

The allow list can be found by first opening Docker Desktop. `Settings -> Resources -> File Sharing` and add the path to the Neosync repository.

If you don't want to do this, the volume mappings can be removed from the compose file, or by removing the PVC for Tilt.
This comes at a negative of the local database not surviving restarts.

### Setup with Tilt

Step 1 is to ensure that the `kind` cluster is up and running along with its registry.
This can be manually created, or done simply with `ctlptl`.
The cluster is declaratively defined [here](./tilt/kind/cluster.yaml)

The below command invokes the cluster-create script that can be found [here](./tilt/scripts/cluster-create.sh)

```
make cluster-create
```

After the cluster has been successfully created, `tilt up` can be run to start up `neosync`.
Refer to the top-level [Tiltfile](./Tiltfile) for a clear picture of everything that runs.
Each dependency in the `neosync` repo is split into sub-Tiltfiles so that they can be run in isolation, or in combination with other sub-resources more easily.

Once everything is up and running, the app can be accessed locally at [http://localhost:3000](http://localhost:3000).

## Analytics

Posthog is used to capture analytics.

Today, they are only captured in a very minimal sense within Neosync app. Eventually this will be extended to the CLI.

You can see what information is captured by checking out the [posthog-provider](https://github.com/nucleuscloud/neosync/blob/main/frontend/apps/web/components/providers/posthog-provider.tsx) component that wraps each page's React components.

Analytics are used simply to get a better view into how people use Neosync.

### Disabling Analytics

If allowing Neosync to capture analytics is not desired, simply remove the `POSTHOG_KEY` from the environment, or disable analytics via the `NEOSYNC_ANALYTICS_ENABLED=false` environment variable.

## Resources

Some resources to help you along the way:

- [Docs](https://docs.neosync.dev) for comprehensive documentation and guides: Note these are still a work in progress.
- [Discord](https://discord.gg/HwrxVfNk) for discussion with the community and Neosync team
- [X](https://x.com/neosynccloud) for the latest updates
<!-- - [Blog](https://neosync.com/blog) for insights, articles and more
- [Roadmap](https://neosync.dev/roadmap) for future development -->

## Contributing

We love contributions big and small. Here are just a few ways that you can contribute to Neosync.

<!-- - Vote on features or get early access to beta functionality in our [roadmap](https://neosync.dev/roadmap) -->

- Join our [Discord](https://discord.gg/HwrxVfNk) channel and ask us any questions there
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
