# Notes on Generating OSS Licenses

## Generate Neosync CA with ED25519

This generates a private key with a password

```console
openssl genpkey -algorithm ed25519 -out neosync_ee_ca.key -aes256
```

## Generate Neosync Pub Key

```console
openssl pkey -in neosync_ee_ca.key -pubout -out neosync_ee_pub.pem
```

## Sign a License File

Signs a file with a provided secret key and generates a signature file

```console
openssl pkeyutl -sign -inkey neosync_ee_ca.key -out license.sig -rawin -in license.json
```

## Verify a License File

Verifies a file with a provided public key and the accompanying signature file

```console
openssl pkeyutl -verify -pubin -inkey neosync_ee_pub.pem -rawin -in license.json -sigfile license.sig
```

## Generate a new License

Refer to the `licenseContents` structure in `internal/ee/license/license.go` for the current makeup of an EE License.

To generate a new `EE_LICENSE` env var key for a new enterprise customer, do the following:

1. Create a license.json file on disk that follows the `licenseContents` struct
2. Ensure the Neosync EE CA private key is accessible on disk
3. Run the following from the root of the repo: `./scripts/gen-cust-license.sh <path>/neosync_ee_ca.key license.json | pbcopy`

This will generate a base64 encoded value stored to the clipboard that follows the `licenseFile` structure.
This may be safely given to an EE customer where they can then plug it in to their locally deployed version of Neosync via the `EE_LICENSE` environment variable.
