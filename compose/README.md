# Compose

Alternative Compose files that can be used to stand up additional services, or alternate versions of Neosync

## compose-auth.yml

Everything in the root `compose.yml` but it also enables authentication through the use of Keycloak.

## compose-auth-prod.yml

This is the same as `compose-auth.yml` but it uses pre-packaged Docker images.
This is a good compose file to run if you just want to try out Neosync (with auth) and not worry about a build environment. All that is needed is Docker!

## compose-prod.yml

This is the same as `compose.yml` (or `compose-auth.yml` minus auth) but it uses pre-packaged Docker images.
This is a good compose file to run if you just want to try out Neosync (without auth) and not worry about a build environment. All that is needed is Docker!

## compose-db.yml

Stands up two separate postgres databases that can be used for testing Neosync.
They are not initialized with any data and it all must be created from scratch.

They expect the `neosync-network` docker network to exist.
