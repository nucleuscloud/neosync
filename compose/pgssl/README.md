# Postgres w/ SSL

This folder is for testing out Postgres Client Certificates.

Root, Server, and Client certificates have been generated and are located in the [./certs](./certs/) folder.

The certificates in this folder are meant for localhost and development use only! They were generated specifically for testing out this feature.

The compose file will bring up a new postgres container with SSL turned on.
This database can be configured in Neosync, and the root+client certs can be added to the connection configuration to further test out a database that uses SSL certificates.
