---
title: Docker Compose
id: docker-compose
hide_title: false
slug: /deploy/docker-compose
---

## Deploying with Docker Compose

A `compose.yml` file is provided in the `compose` folder to easily and quickly get Neosync up and running without having to build any dependencies. All that's needed is `docker`.

This compose file is not tailored for development environments as it uses pre-baked docker images that are intended for use in production environments.

There is a companion compose file found in the `temporal` folder that should be run prior to the main `compose.yml` to stand up an instance of Temporal and all of its dependencies.

To simplify the setup even further, a compose file is provided that does not stand up authentication for Neosync. Check out the section below to stand up Neosync with auth.

```sh
$ make compose-up
```

Once all of the containers come online, the app is now routable via [http://localhost:3000](http://localhost:3000).

## Deploy with Docker Compose and Authentication

Neosync provides an auth friendly compose file that will stand up Neosync in auth-mode with Keycloak.

```sh
$ make compose-auth-up
```

Keycloak comes default with two clients that allow the app and cli to login successfully.

On first boot up, Keycloak will assert itself with the provided realm Neosync realm.

When navigating to Neosync for the first time, you'll land on the Keycloak sign-in page. It is easy to create an account simply by going through the register flow.
This will persist restarts due to the postgres volume mapping. If you wish to start over, simply delete the neosync docker volume to reset your database to a fresh state.
