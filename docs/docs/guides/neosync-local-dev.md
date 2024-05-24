---
title: Developing Neosync Locally
description: Learn how to develop with Neosync Open Source locally in order to get up to speed with how Neosync works
id: neosync-local-dev
hide_title: false
slug: /guides/neosync-local-dev
---

## Introduction

This section goes into detail on each tool that is used for developing with Neosync locally.

## Neosync DevContainer

Neosync has a pre-published [devcontainer](https://containers.dev/) that can be used to easily get a working Neosync dev environment.
This container comes pre-packaged with all of the tools needed for developing Neosync, and works with Tilt or Compose, or Bare Metal setups.

## Setup with Compose

### Pre-requisites

- Golang >=1.22
- Docker Compose >=2.26

### Setup

The docker compose environment runs almost entirely by itself. However, the Golang services (backend, worker), must be built on the host system.

Each service can be built simply by `cd`ing into each directory (`backend`, `worker`) and running the command `make dbuild`. This builds the service specifically for Linux.

Running `make compose-dev-up` will stand up Neosync and it's dependencies.
**Currently there is a limitation with devcontainers where this command must be run via `sudo`.**

The frontend, backend, and worker will be in watch mode, which will cause them to look for changes and reload. The frontend will reload on save, where as the backend/worker will reload on the next `make dbuild` run.

```sh
cd backend && make dbuild
cd worker && make dbuild
```

Once everything is up and running, The app can be accessed locally at [http://localhost:3000](http://localhost:3000).

### Running Compose with Authentication

This will stand up Keycloak with a pre-configured realm that will allow logging in to Neosync with a standard username and password, completely offline!

```sh
make compose-dev-auth-up
```

To stop, run:

```sh
make compose-dev-auth-down
```

## Setup with Tilt

Developing with Kubernetes via Tilt is also an option, however it is a bit more setup and is heavier. The benefits of this are that it allows you to develop more closely to what a k8s production environment could look like.d

### Docker Desktop

If using Docker Desktop, the host file path to the `.data` folder will need to be added to the File Sharing tab.

The allow list can be found by first opening Docker Desktop. `Settings -> Resources -> File Sharing` and add the path to the Neosync repository.

If you don't want to do this, the volume mappings can be removed by removing the PVC for Tilt.
This comes at a negative of the local database not surviving restarts.

### Cluster Setup

Step 1 is to ensure that the `kind` cluster is up and running along with its registry.
This can be manually created, or done simply with `ctlptl`.
The cluster is declaratively defined [here](https://github.com/nucleuscloud/neosync/tree/main//tilt/kind/cluster.yaml)

The below command invokes the cluster-create script that can be found [here](https://github.com/nucleuscloud/neosync/tree/main//tilt/scripts/cluster-create.sh)

```
make cluster-create
```

After the cluster has been successfully created, `tilt up` can be run to start up `neosync`.
Refer to the top-level [Tiltfile](https://github.com/nucleuscloud/neosync/tree/main//Tiltfile) for a clear picture of everything that runs.
Each dependency in the `neosync` repo is split into sub Tilt files so that they can be run in isolation, or in combination with other sub-resources more easily.

Once everything is up and running, the app can be accessed locally at [http://localhost:3000](http://localhost:3000).

## Developing on Bare Metal

You can develop Neosync totally on bare metal. Every service supports a .env file along with environment specific .env overrides.
This way of developing isn't really used today as we've invested heavily in developing within containerized environments to be more closely aligned with a production environment.

## Tools

This section contains a flat list of the tools that are used to develop Neosync and why.

Detailed below are the main dependencies are descriptions of how they are utilized:

### Kubernetes

If you're choosing to develop in a Tilt environment, this section is more important as it contains all of the K8s focused tooling.

Tilt is a great tool that is used to automate the setup of a Kubernetes cluster. There are multiple `Tiltfile`'s throughout the code, along with a top-level one that is used to inject all of the K8s manifests to setup Neosync inside of a K8s cluster.

This enables fast development, locally, while closely mimicking a real production environment.

- [kind](https://github.com/kubernetes-sigs/kind)
  - Kubernetes in Docker. We use this to spin up a slim kubernetes cluster that deploys all of the `neosync` resources.
- [tilt](https://github.com/tilt-dev/tilt)
  - Allows us to define our development environment as code.
- [ctlptl](https://github.com/tilt-dev/ctlptl)
  - CLI provided by the Tilt-team to make it easy to declaratively define the kind cluster that is used for development
- [kubectl](https://github.com/kubernetes/kubectl)
  - Allows for observability and management into the spun-up kind cluster.
- [kustomize](https://github.com/kubernetes-sigs/kustomize)
  - yaml template tool for ad-hoc patches to kubernetes configurations
- [helm](https://github.com/helm/helm)
  - Kubernetes package manager. All of our app deployables come with a helm-chart for easy installation into kubernetes
- [helmfile](https://github.com/helmfile/helmfile)
  - Declaratively define a helmfile in code! We have all of our dev charts defined as a helmfile, of which Tilt points directly to.

### Go + Protobuf

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

### Npm/Nodejs

- [Node/Npm](https://nodejs.org/en)
  - Used to run the app, along with Nextjs.

All of these tools can be easily installed with `brew` if on a Mac.
Today, `sqlc` and `buf` don't need to be installed locally as we exec docker images for running them.
This lets us declare the versions in code and docker takes care of the rest.

## Brew Install

Each tool above can be straightforwardly installed with brew if on Linux/MacOS

```
brew install kind tilt-dev/tap/tilt tilt-dev/tap/ctlptl kubernetes-cli kustomize helm helmfile go sqlc buf golangci-lint node
```
