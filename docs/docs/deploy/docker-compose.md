---
title: Docker Compose
id: docker-compose
hide_title: false
slug: /deploy/docker-compose
---

## Deploying with Docker Compose

A `compose.yml` file is provided in the root of the repository to easily and quickly get Neosync up and running.
This compose file is currently tailored for localhost/development environments, but should be easily adaptable to production environments.

There is a companion compose file found in the `temporal` folder that should be run prior to the main `compose.yml` to stand up a development instance of Temporal and all of it's dependencies.

Due to being a development focused setup, both Temporal and Neosync are started up without Authentication enabled. This is done to simplify configuration and standup time.
Database volumes for both Tempral and Neosync are mapped to the `.data` folder inside of the repository. Note: If using Docker Desktop, this folder will have to be added to the list of allowed file system mappings prior to running docker compose.

```sh
$ docker compose -f temporal/compose.yml up -d
$ docker compose up -d
```

Once all of the containers come online, the app is now routable via `http://localhost:3000`.

## Deploy with Docker Compose and Authentication

Neosync provides an auth friendly compose file that will stand up Neosync in auth-mode with Keycloak.

```sh
$ docker compose -f temporal/compose.yml up -d
$ docker compose -f compose/compose-auth.yml up -d
```

It comes default with two clients that allow the app and cli to login successfully.

On first boot up, Keycloak will assert itself with the provided realm Neosync realm.

When navigating to Neosync for the first time, you'll land on the Keycloak signin page. It is easy to create an account simply by going through the register flow.
This will persist restarts due to the postgres volume mapping. If you wish to start over, simply delete your `.data/neosync-postgres` folder to do so.
