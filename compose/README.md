# Compose

Alternative Compose files that can be used to stand up additional services, or alternate versions of Neosync

## compose-auth.yml

Everything in the root `compose.yml` but it also enables authentication through the use of Keycloak.

## compose-db.yml

Stands up two separate postgres databases that can be used for testing Neosync.
They are not initialized with any data and it all must be created from scratch.

They expect the `neosync-network` docker network to exist.
