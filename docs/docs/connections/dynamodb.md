---
title: DynamoDB
description: Amazon DynamoDB is a fully managed proprietary NoSQL database offered by Amazon.com as part of the Amazon Web Services portfolio.
id: dynamodb
hide_title: false
slug: /connections/dynamodb
# cSpell:words dummyid,dummysecret
---

## Introduction

DynamoDB offers a fast persistent key–value datastore with built-in support for replication, autoscaling, encryption at rest, and on-demand backup among other features. It is one of the most highly requested database connections to be added to Neosync.

If you are interested in using DynamoDB but don't see a feature that is required for you to use it, please reach out to us on Discord!

## Configuring DynamoDB

There are a few different methods of giving Neosync access to your DynamoDB instance. This section will talk through them and also detail the necessary IAM permissions required for Neosync to function.

### IAM Role Access

Neosync supports being given an IAM Role along with an External ID.

If configuring DynamoDB via NeosyncCloud, this is the recommended approach over using raw Access Credentials that don't expire. Neosync will assume this role only during active syncs or any time data is requested via the frontend and does not store those credentials in any way.

### AWS Access Credentials

Neosync supports configuring static access credentials, or if you are just trying things out, you can configure temporary credentials with a session token.
This is fine for testing, but is not recommended for a production setup.

Static access credentials are heavily discouraged as they do not expire.

### Self-Hosted with AWS

If you are self-hosting Neosync, it's recommended to provide Neosync with minimal configuration via a Neosync Connection and to instead attach an IAM role to the process directly.
This way the running neosync-api and neosync-worker are able to natively have necessary permissions to read/write DynamoDB tables.

For example, if hosting an EKS cluster, it's recommended to attach an IAM IRSA role to the Neosync deployments with the policies detailed below instead of configuring them directly in the application.
You'll still need to create the DynamoDB connections inside of Neosync, but the configuration will essentially be empty.

## NeosyncCloud Trust Policy

The NeosyncCloud principal is: `arn:aws:iam::243317024749:root`, which will allow our cloud services to communicate with your DynamoDB instance.
Be sure to update the `sts:ExternalId` property with the external id that you've configured with the role.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::243317024749:root"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "<external-id>"
        }
      }
    }
  ]
}
```

Continue below to see what permissions are necessary for readonly access as well as readwrite.

## Configuring a Policy

The best way to talk about a policy is to give an example scenario.

Let's say there are two DynamoDB tables, both in the same AWS Account: `ProdDb` and `StageDb`.
We want to build two policies that will be attached to a role that will grant readonly access to `ProdDb` and readwrite access to `StageDb`.

### Readonly Policy

This policy will grant readonly access to the `ProdDb` table.
This also grants `ListTables`, which is needed by the frontend for showing Permissions and to allow mapping tables while configuring a Neosync job.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:BatchGetItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:DescribeTable",
        "dynamodb:PartiQLSelect"
      ],
      "Resource": ["arn:aws:dynamodb:<region>:<accountId>:table/ProdDb"]
    },
    {
      "Effect": "Allow",
      "Action": ["dynamodb:ListTables"],
      "Resource": ["*"]
    }
  ]
}
```

### ReadWrite Policy

This policy will grant readwrite access to the `StageDb` table.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:BatchGetItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:DescribeTable",
        "dynamodb:PartiQLSelect"
      ],
      "Resource": ["arn:aws:dynamodb:<region>:<accountId>:table/StageDb"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:BatchWriteItem",
        "dynamodb:PartiQLInsert",
        "dynamodb:PartiQLUpdate"
      ],
      "Resource": ["arn:aws:dynamodb:<region>:<accountId>:table/StageDb"]
    },
    {
      "Effect": "Allow",
      "Action": ["dynamodb:ListTables"],
      "Resource": ["*"]
    }
  ]
}
```

## Neosync Cloud Region

Neosync Cloud currently runs in `us-west-2` region. If you've using the cloud platform and are running your tables in a different region, be sure to fill out the region field for the Neosync Connection to ensure that Neosync looks in the right place for your DynamoDB tables.

## Sync Job Mapping Configuration

When configuring a sync with DynamoDB, there are a few items to consider when choosing what should be added to mapped columns.

As of today, all records in a table are scanned and brought over to the destination (to reduce this, see subsetting below).

Fill out the mapping form and choose a transformer to allow specific keys to be transformed during this sync process.

### Mapping of unmapped keys

You also have the ability to configure default transformers for any _unmapped_ key.
This will automatically transform values by their type, unless it is a known key (i.e. it has been added directly to the job mappings table.)

Neosync has selected sensible defaults for each primitive type, but if it's desired to pass every value through except for known mapped keys, this can be configured as well.

Neosync by default will not map primary key columns as this would result in new records being created every time.
As such, all primary key columns are by default marked as passthrough, unless otherwise configured as a known mapping.

### How to map keys with DynamoDB

DynamoDB stores values in JSON format, with the type of the value encoded directly into the stored JSON array.

Example:

```json
{
  "Id": {
    "S": "111"
  },
  "Name": {
    "S": "Nick"
  },
  "Age": {
    "N": "32"
  },
  "Pets": {
    "M": {
      "Judith": {
        "M": {
          "Type": {
            "S": "Cat"
          },
          "Age": {
            "N": "4"
          }
        }
      }
    }
  }
}
```

These types get transformed and converted into language type primitives.

In other words, the object is flattened and the types are converted into Go types.
So to work with this object, the key to be provided should omit the types from the path.

If I wanted to anonymize the name of my cat, I would provide the key: `Pets.Judith.Age`, _not_ `Pets.M.Judith.M.Age`

### Mapping Limitations

The main limitation with the mapping scheme in its correct form is the inability to provide mapping keys for specific array indexes.

It is currently recommended to provide the key directly to the array itself and configure a `TransformJavaScript` or `GenerateJavaScript` transformer to convert the array using JS instead of an off the shelf transformation.
All of Neosync's transformer functions are available and usable within the JS transformers, so the sky is the limit for what you can do there.

### Data Type Gotchas

The sync processor is written in Go. Go does not have a native Set type. To work with DynamoDB's native set, these are converted into Go slices.
Metadata for the original type is retained and the transformed data is converted back into a DynamoDB Set prior to being written.

## Dataset Subsetting

Neosync by default scans an entire table during a sync. This can be reduced via a table subset.
Neosync uses PartiQL for DB scanning as it offers a much simpler configuration interface for writing such queries.

The traditional approach is very type heavy, where PartiQL offers a more automatic approach to handling this type conversion, at least for simple cases.

The default query looks basically like this: `SELECT * FROM <table>`.

IF a subset is configured, a where clause is tacked on along with your provided query. Example: `SELECT * FROM <table> WHERE Id = '111'`.

If this type of subsetting is not sufficient for your usecase, please reach out to us on Discord.

## Trying out DynamoDB offline

Even though DynamoDB is AWS proprietary, they ship a [Docker image](https://hub.docker.com/r/amazon/dynamodb-local) that can be used to try out DynamoDB in an offline, self-hosted format.

Neosync provides a Docker [compose.yml](https://github.com/nucleuscloud/neosync/blob/main/compose/compose-db-dynamo.yml) that can be used in conjunction with our main `compose` or dev compose to stand up two local instances inside of the `neosync-network` docker network. In order to enable this in the `compose.dev.yaml` file, you just need to uncomment the `  - path: ./compose/compose-db-dynamo.yml` line, like so:

```yaml
include:
  - path: ./compose/temporal/compose.yml
    env_file:
      - ./compose/temporal/.env
  - path: ./compose/compose-db.yml
  # - path: ./compose/compose-db-mysql.yml
  # - path: ./compose/compose-db-mongo.yml
  - path: ./compose/compose-db-dynamo.yml
  # - path: ./compose/compose-db-mssql.yml
```

Both are not needed and could definitely be trimmed down to one, but they can be used as two discrete AWS Accounts / Two totally different DynamoDB servers.

These databases do require basic access credentials, so be sure when accessing these (inside or outside of Neosync), you provide the same access key, as otherwise DynamoDB treats it as a separate user and will not return the same data.

In order to set up Dynamo DB locally, install the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html). Then create/update the AWS credentials file with a `dummy` profile, like so:

```console
cat ~/.aws/credentials
[dummy]
aws_access_key_id = dummyid
aws_secret_access_key = dummysecret
```

Then using the AWS CLI, create a table using this command:

```console
aws dynamodb create-table --table-name ExampleTable \
  --attribute-definitions AttributeName=Id,AttributeType=S \
  --key-schema AttributeName=Id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --endpoint-url http://localhost:8009 --profile dummy \
  --region us-west-2
```

The `endpoint-url` here corresponds to the `test-stage-db-dynamo` container which is at port `8009`. To set up another db, just update the port in in the `endpoint-url`.

Once you've created the table in the database, then head over to Neosync and create a Dynamo DB connection. In order to connect to your local Dynamo DB instance, you'll need to fill out:

- Connection Name - ex. dynamo-stage
- Access Key ID - ex. dummyid
- AWS Secret Key - ex. dummysecret
- AWS Advanced Configuration -> AWS Region - us-west-2
- AWS Advanced Configuration -> Custom Endpoint - http://test-stage-db-dynamo:8000

Then click Test Connection in order to verify that your connection is set up correctly.
