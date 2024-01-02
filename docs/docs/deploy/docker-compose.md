---
title: Docker Compose
id: docker-compose
hide_title: false
slug: /deploy/docker-compose
---

## Deploying with Docker Compose

A `compose.yml` file is provided in the `compose` folder to easily and quickly get Neosync up and running without having to build any dependencies. All that's needed is `docker`.

This compose file is not tailed for development environments as it uses pre-baked docker images that are intended for use in production environments.

There is a companion compose file found in the `temporal` folder that should be run prior to the main `compose.yml` to stand up an instance of Temporal and all of it's dependencies.

To simplify the setup even further, a compose file is provided that does not stand up authentication for Neosync. Check out the section below to stand up Neosync with auth.

Database volumes for both Tempral and Neosync are mapped to the `.data` folder inside of the repository. Note: If using Docker Desktop, this folder will have to be added to the list of allowed file system mappings prior to running docker compose.

```sh
$ docker compose -f temporal/compose.yml up -d
$ docker compose -f compose/compose-prod.yml up -d
```

Once all of the containers come online, the app is now routable via [http://localhost:3000](http://localhost:3000).

## Deploy with Docker Compose and Authentication

Neosync provides an auth friendly compose file that will stand up Neosync in auth-mode with Keycloak.

```sh
$ docker compose -f temporal/compose.yml up -d
$ docker compose -f compose/compose-auth-prod.yml up -d
```

Keycloak comes default with two clients that allow the app and cli to login successfully.

On first boot up, Keycloak will assert itself with the provided realm Neosync realm.

When navigating to Neosync for the first time, you'll land on the Keycloak signin page. It is easy to create an account simply by going through the register flow.
This will persist restarts due to the postgres volume mapping. If you wish to start over, simply delete your `.data/neosync-postgres` folder to do so.

## Developing Neosync with Compose

Check out the README at the root of the Neosync Github repository to learn more about how to development Neosync using compose and the dev-focused compose files.
