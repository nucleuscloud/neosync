---
title: System Transformers
description: Learn about Neosync's system transformers that come out of the box and help anonymize data and generate synthetic data
id: system
hide_title: false
slug: /transformers/system
# cSpell:words luhn,huntingon,dach,sporer,rodriguezon,rega,jdoe,lsmith,aidan,kunze,littel,johnsonston
---

## Introduction

Neosync comes out of the box with 40+ system transformers in order to give you an easy way to get started with anonymizing or generating data. There are two kinds of system transformers:

1. Generators - these begin with `generate_` and do not anonymize existing data. They only generate net-new synthetic data.
2. Transformers - these begin with `transform_` and can anonymize existing data or generate new-new synthetic data based on how you configure it.

## Reference

| Name                                                                              | Type    | Code                                                                                                                                                 | Description                                                                                                                      |
| --------------------------------------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| [Generate Categorical](/transformers/system#generate-categorical)                 | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_categorical.go)                               | Randomly selects a value from a defined set of categorical values                                                                |
| [Generate Email](/transformers/system#generate-email)                             | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_email.go)                                     | Generates a new randomized email address.                                                                                        |
| [Transform Email](/transformers/system#transform-email)                           | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_email.go)                                    | Transforms an existing email address.                                                                                            |
| [Generate Boolean](/transformers/system#generate-boolean)                         | boolean | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_bool.go)                                      | Generates a boolean value at random.                                                                                             |
| [Generate Card Number](/transformers/system#generate-card-number)                 | int64   | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_card_number.go)                               | Generates a card number.                                                                                                         |
| [Generate City](/transformers/system#generate-city)                               | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_city.go)                                      | Randomly selects a city from a list of predefined US cities.                                                                     |
| [Generate Default](/transformers/system#generate-default)                         | any     | [Code](https://github.com/nucleuscloud/neosync/blob/4e459151080109ffa0c5b0d9937f4f20fecfa6c6/worker/pkg/workflows/datasync/activities/activities.go) | Applies the predefined default value of a column as specified in the SQL database schema                                         |
| [Generate E164 Phone Number](/transformers/system#generate-e164-phone-number)     | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_e164_phone.go)                                | Generates a Generate phone number in e164 format.                                                                                |
| [Generate First Name](/transformers/system#generate-first-name)                   | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_first_name.go)                                | Generates a random first name.                                                                                                   |
| [Generate Float64](/transformers/system#generate-float64)                         | float64 | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_float.go)                                     | Generates a random float64 value.                                                                                                |
| [Generate Full Address](/transformers/system#generate-full-address)               | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_full_address.go)                              | Randomly generates a street address.                                                                                             |
| [Generate Full Name](/transformers/system#generate-full-name)                     | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_full_name.go)                                 | Generates a new full name consisting of a first and last name.                                                                   |
| [Generate Gender](/transformers/system#generate-gender)                           | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_gender.go)                                    | Randomly generates one of the following genders: female, male, undefined, nonbinary.                                             |
| [Generate Int64 Phone Number](/transformers/system#generate-int64-phone-number)   | int64   | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_164_phone.go)                                 | Generates a new phone number of type int64 with a default length of 10.                                                          |
| [Generate Random Int64](/transformers/system#generate-random-int64)               | int64   | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_int.go)                                       | Generates a random integer value with a default length of 4 unless the Integer Length or Preserve Length parameters are defined. |
| [Generate Last Name](/transformers/system#generate-last-name)                     | int64   | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_last_name.go)                                 | Generates a random last name.                                                                                                    |
| [Generate SHA256 Hash](/transformers/system#generate-sha256-hash)                 | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_sha256hash.go)                                | SHA256 hashes a randomly generated value.                                                                                        |
| [Generate SSN](/transformers/system#generate-ssn)                                 | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_ssn.go)                                       | Generates a completely random social security numbers including the hyphens in the format `xxx-xx-xxxx`.                         |
| [Generate State](/transformers/system#generate-state)                             | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_state.go)                                     | Randomly selects a US state and returns the two-character state code.                                                            |
| [Generate Street Address](/transformers/system#generate-street-address)           | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_street_address.go)                            | Randomly generates a street address.                                                                                             |
| [Generate String Phone Number](/transformers/system#generate-string-phone-number) | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_string_phone.go)                              | Generates a Generate phone number and returns it as a string.                                                                    |
| [Generate Random String](/transformers/system#generate-random-string)             | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_string.go)                                    | Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length parameter are defined.       |
| [Generate Unix Timestamp](/transformers/system#generate-unix-timestamp)           | int64   | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_unixtimestamp.go)                             | Randomly generates a Unix timestamp.                                                                                             |
| [Generate Username](/transformers/system#generate-username)                       | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_username.go)                                  | Randomly generates a username                                                                                                    |
| [Generate UTC Timestamp](/transformers/system#generate-utc-timestamp)             | time    | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_utctimestamp.go)                              | Randomly generates a UTC timestamp.                                                                                              |
| [Generate UUID](/transformers/system#generate-uuid)                               | uuid    | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_uuid.go)                                      | Generates a new UUIDv4 id.                                                                                                       |
| [Generate Zipcode](/transformers/system#generate-zipcode)                         | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/generate_zipcode.go)                                   | Randomly selects a zip code from a list of predefined US zipcodes.                                                               |
| [Transform E164 Phone Number](/transformers/system#transform-e164-phone-number)   | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_e164_phone.go)                               | Transforms an existing E164 formatted phone number.                                                                              |
| [Transform First Name](/transformers/system#transform-first-name)                 | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_first_name.go)                               | Transforms an existing first name                                                                                                |
| [Transform Float64](/transformers/system#transform-float64)                       | float64 | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_float.go)                                    | Transforms an existing float value.                                                                                              |
| [Transform Full Name](/transformers/system#transform-full-name)                   | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_full_name.go)                                | Transforms an existing full name.                                                                                                |
| [Transform Int64 Phone Number](/transformers/system#transform-int64-phone-number) | int64   | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_int_phone.go)                                | Transforms an existing phone number that is typed as an integer                                                                  |
| [Transform Int64](/transformers/system#transform-int64)                           | int64   | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_int.go)                                      | Transforms an existing integer value.                                                                                            |
| [Transform Last Name](/transformers/system#transform-last-name)                   | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_last_name.go)                                | Transforms an existing last name.                                                                                                |
| [Transform Phone Number](/transformers/system#transform-phone-number)             | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_phone.go)                                    | Transforms an existing phone number that is typed as a string.                                                                   |
| [Transform String](/transformers/system#transform-string)                         | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_string.go)                                   | Transforms an existing string value.                                                                                             |
| [Transform Character Scramble](/transformers/system#transform-character-scramble) | string  | [Code](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/transform_character_scramble.go)                       | Transforms an existing string value by scrambling the characters while maintaining the format.                                   |
| [Passthrough](/transformers/system#passthrough)                                   | string  | [Code](https://github.com/nucleuscloud/neosync/blob/4e459151080109ffa0c5b0d9937f4f20fecfa6c6/worker/pkg/workflows/datasync/activities/activities.go) | Passes the input value through to the destination with no changes.                                                               |
| [Null](/transformers/system#null)                                                 | string  | [Code](https://github.com/nucleuscloud/neosync/blob/4e459151080109ffa0c5b0d9937f4f20fecfa6c6/worker/pkg/workflows/datasync/activities/activities.go) | Inserts a `null` string instead of the source value.                                                                             |

### Generate Categorical\{#generate-categorical}

The generate categorical transformer allows a user to define a list of comma-separated categorical string values that Neosync will randomly sample.

For example, the following values:
`test, me, please`

Could produce the following output value:
`me`

Or:
`please`

**Configurations**

| Name       | Description                                                                          | Default         | Example Output |
| ---------- | ------------------------------------------------------------------------------------ | --------------- | -------------- |
| Categories | list of comma-separated categorical string values that Neosync will randomly sample. | `value1,value2` | `value1`       |

**Examples**

| Categories | Example Input    | Example Output |
| ---------- | ---------------- | -------------- |
| false      | `red,blue,green` | `green`        |
| false      | `red,blue,green` | `red`          |

### Generate Email\{#generate-email}

The generate email transformer generates a new randomized email address in the format:

`<username>@<domain>.<top-level-domain>`

By default, the transformer randomizes the username, domain and top-level domain while always preserving the email format by retaining the `@` and `.` characters.

For example, the following input value:
`john@acme.com`

Would produce the following output value:
`ytvub873@gmail.com`

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output        |
| --------------------- |
| `f98723uh@gmail.com`  |
| `9cd@msn.com`         |
| `mDy@gmail.edu `      |
| `fweq23f@hotmail.com` |

### Transform Email\{#transform-email}

The transform email transformer can anonymize an existing email address or generate a new randomized email address in the format:

`<username>@<domain>.<top-level-domain>`

By default, the transformer randomizes the username, domain and top-level domain while always preserving the email format by retaining the `@` and `.` characters.

For example, the following input value:
`john@acme.com`

Would produce the following output value:
`ytvub873@gmail.com`

**Configurations**

Depending on your logic, you may want to configure the output email. The transform email transformer has the following configurations:

| Name             | Description                                                                                                                                                                                                               | Default | Example Input           | Example Output                               |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ----------------------- | -------------------------------------------- |
| PreserveDomain   | Carries over the input domain and top-level to the output email address.                                                                                                                                                  | false   | `john@gmail.com`        | `fjf2903@gmail.com`                          |
| PreserveLength   | Ensures the output email address is the same length as the input email address.                                                                                                                                           | false   | `john@gmail.com`        | `hw98@gmail.edu `                            |
| Excluded Domains | Specify a list of comma separated domains to exclude from transformation logic.                                                                                                                                           | empty   | `gmail.com,hotmail.com` | `hw98@gmail.edu `                            |
| Email Type       | Specify how the email should be transformed. Today the options are uuid v4 and fullname. Fullname is not guaranteed to be unique, although it has high cardinality. For unique transformers, it's advised to use uuid v4. | uuid v4 | `foo@gmail.com`         | `d62bcb6a98ee4316a078852efda21c3d@gmail.com` |

**Examples**

There are several ways you can mix-and-match configurations to get different potential email formats. Note that the Excluded Domains List pertains to the transformed emails. If the PreserverDomain is true and there is a domain in the Excluded Domains list, then the default logic would be to **not** preserve the domain but since we are excluding those domains from the transformation logic via the exclude list, then those domains **will** be excluded.

Here are some possible combinations:

| PreserveLength | PreserveDomain | Exclude Domains | Example Input        | Example Output        |
| -------------- | -------------- | --------------- | -------------------- | --------------------- |
| false          | true           | `gmail.com`     | `evis@gmail.com`     | `f9nkeuh@ergwer.wewe` |
| true           | false          | `gmail.com`     | `f98723uh@gmail.com` | `f98723uh@gmail.com`  |
| true           | true           | `gmail.com`     | `evis@gmail.com`     | `f98723uh@weefw.wefw` |
| false          | false          | `gmail.com`     | `evis@gmail.com`     | `f98723uh@gmail.com`  |

### Generate Boolean\{#generate-boolean}

The generate boolean transformer randomly generates a boolean value.

**Configurations**

There are no configurations for the random boolean transformer.

**Examples**

Here are some examples of what an output random boolean value may look like.

| Example Output |
| -------------- |
| true           |
| false          |

### Generate Card Number\{#generate-card-number}

The generate card number transformer generates a new card number that is [luhn valid.](https://en.wikipedia.org/wiki/Luhn_algorithm)

By default, the card number transformer generates a random 16 digit card number that is _not_ luhn valid. If you want luhn validation, please set the luhn-check config to `true`.

**Configurations**

Depending on your validations, you may want to configure the output card number. The card number transformer has the following configurations:

| Name       | Description                                           | Default | Example Output   |
| ---------- | ----------------------------------------------------- | ------- | ---------------- |
| Valid Luhn | Generate a card number that can pass luhn validation. | true    | 6621238134011099 |

**Examples**

Here are some examples of an output card number:

| Valid Luhn | Example Output   |
| ---------- | ---------------- |
| false      | 3627893345931223 |
| true       | 6621238134011099 |

### Generate City\{#generate-city}

The generate city transformers generates a randomly selected US city. You can see the complete list of cities that are available to be randomly selected [here](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/addresses.json).

If you'd rather not get back a real city value, you can use the [Generate Random String Transformer](/transformers/system#generate-random-string) to generate a random string value.

**Configurations**

There are no configurations for the city transformer.

**Examples**

Here are some examples of what an output City value may look like.

| Example Output |
| -------------- |
| Louisville     |
| Manchester     |

### Generate Default\{#generate-default}

The generate default transformer automatically applies the predefined default value of a column as specified in the SQL database schema. This means whenever new data is added without specifying a value for this column, the system will insert the default value set for that column in the database.

### Generate E164 Phone Number\{#generate-e164-phone-number}

The generate e164 phone transformer generates a new international phone number including the + sign. By default, the generate e164 transformer generates a random phone number with no hyphens in the format:

`+34567890123`

TPhone numbers also vary in length with some international phone numbers reaching up to 15 digits in length. You can set the range of the value by setting the `min` and `max` params. Here is more information on the [e164 format](https://www.twilio.com/docs/glossary/what-e164). This transformer also has a min of 9 and a max of 15. If you want to generate a number that is longer than that, you can use the [Generate Random int64 transformer](/transformers/system#generate-random-int64)

**Configurations**

Depending on your validations, you may want to configure the output e164 phone number:

| Name | Description                                   | Default | Example Output |
| ---- | --------------------------------------------- | ------- | -------------- |
| Min  | Min is the lower range of the a length value. | 9       | +29748675      |
| Max  | Max is the upper range of a length value.     | 12      | +2037486752    |

**Examples**

There are several ways you can mix-and-match configurations to get different potential phone number formats. Here are some possible combinations:

| Min | Max | Example Output |
| --- | --- | -------------- |
| 9   | 12  | +5209273239    |
| 10  | 14  | +5209273239    |
| 12  | 12  | +52092732323   |

### Generate First Name\{#generate-first-name}

The generate first name transformer generates a valid first name from a list of predefined first name values. You can see the entire list of first name value [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/first-names.json)

By default, the generate first name transformer randomly picks a first name with a length between 2 and 12.

**Configurations**

There are no configurations for this transformer.

**Examples**

There are several ways you can mix-and-match configurations to get different first name formats. Here are some possible combinations:

| Example Output |
| -------------- |
| Benjamin       |
| Gregory        |
| Ashley         |

### Generate Float64\{#generate-float64}

The generate float transformer generates a random floating point number.

For example:
`32.2432`

This transformer supports both positive and negative float numbers with a max precision of 17.

Go float64 adheres to the IEEE 754 standard for double-precision floating-point numbers.

If this is not sufficient, contact us about creating a generator that supports numbers with higher precision.

**Configurations**

Depending on your validations, you may want to configure the output float. The random float transformer has the following configurations:

| Name          | Description                                                                                           | Default |
| ------------- | ----------------------------------------------------------------------------------------------------- | ------- |
| RandomizeSign | Randomly sets the sign of the value.                                                                  | false   |
| Min           | Specifies the min range for a generated float value.                                                  | 1.00    |
| Max           | Specifies the max range for a generated float value.                                                  | 100.00  |
| Precision     | Sets the precision of the value. For example, a precision of 6 will generate a value such as 42.9812. | 6       |

**Examples**

There are several ways you can mix-and-match configurations to get different potential random float formats.

Note: if randomize sign has been selected, this may cause the generated numbers to exist outside of the configured min/max range.

Here are some possible combinations:

| RandomizeSign | Min    | Max   | Precision | Example Output |
| ------------- | ------ | ----- | --------- | -------------- |
| true          | 2.00   | 35.00 | 4         | -23.84         |
| true          | 4.12   | 19.43 | 3         | 7.28           |
| false         | -30.00 | 2.00  | 8         | -15.54219      |
| false         | -20.00 | 30.00 | 4         | 20.19          |

### Generate Full Address\{#generate-full-address}

The generate full address transformer generates a randomly selected real full address that exists in the United States. You can see the complete list of full addresses that are available to be randomly selected [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/addresses.json)

The full address transformer returns a valid United States address. For example:
`123 Main Street Boston, Massachusetts, 02169`

**Configurations**

There are no configurations for the full address transformer.

**Examples**

Here are some examples of what an output full address value may look like.

| Example Output                                      |
| --------------------------------------------------- |
| 509 Franklin Street Northeast Washington, DC, 20017 |
| 14 Huntingon Street Manchester, CT 06040            |

### Generate Full Name\{#generate-full-name}

The generate full name transformer generates a valid full name from a list of predefined full name values. The generated full name is made from a combination of the [first](/transformers/system#generate-first-name) and [last](/transformers/system#generate-last-name) names transformers.

**Configurations**

There are no configurations for this transformer.

**Examples**

Here are some examples of what an output full name value may look like.

| Example Output  |
| --------------- |
| Edward Marshall |
| Dante Mayer     |
| Gregory Johnson |

### Generate Gender\{#generate-gender}

The generate gender transformer randomly selects a gender value from a predefined list of genders. Here is the list:

| Gender    | Abbreviation |
| --------- | ------------ |
| male      | m            |
| female    | f            |
| nonbinary | n            |
| undefined | u            |

By default, the gender transformer does not abbreviate the gender. If you'd like to return an abbreviated gender, pass in the `abbreviate` config.

**Configurations**

Depending on your validations, you may want to configure the output gender. The gender transformer has the following configurations:

| Name       | Description                                                                    | Default | Example Output |
| ---------- | ------------------------------------------------------------------------------ | ------- | -------------- |
| Abbreviate | Abbreviate will abbreviate the output gender so that it is only one character. | false   | u              |

**Examples**

There are several ways you can mix-and-match configurations to get different full name formats. Here are some possible combinations:

| Abbreviate | Example Output |
| ---------- | -------------- |
| false      | male           |
| true       | f              |
| false      | nonbinary      |

### Generate Int64 Phone Number\{#generate-int64-phone-number}

The generate int64 phone number generates a random 10 digit phone number with a valid US area code and returns it as an `int64` type with no hyphens. If you want to return a `string` or want to include hyphens, check out the [Generate String Phone transformer](/transformers/system#generate-string-phone-number).

For example, the generate int64 phone number transformer could generate the following output value:
`5698437232`

**Configurations**

There are no configurations for this transformer.

**Examples**

Here are some examples of what the output values could look like:

| Example Output |
| -------------- |
| 5209273239     |
| 9378463528     |

### Generate Random Int64\{#generate-random-int64}

The generate random int64 transformer generates a random integer and returns it as a int64 type.

For example:
`6782`

**Configurations**

Depending on your validations, you may want to configure the output float. The random integer transformer has the following configurations:

| Name          | Description                                          | Default |
| ------------- | ---------------------------------------------------- | ------- |
| RandomizeSign | Randomly sets the sign of the generated value.       | false   |
| Min           | Specifies the min range for a generated int64 value. | 1       |
| Max           | Specifies the max range for a generated int64 value. | 40      |

**Examples**

There are several ways you can mix-and-match configurations to get different potential random integer formats.

Note: if randomize sign has been selected, this may cause the generated numbers to exist outside of the configured min/max range.

Here are some possible combinations:

| RandomizeSign | Min | Max | Example Output |
| ------------- | --- | --- | -------------- |
| true          | 10  | 40  | -23            |
| true          | -20 | -3  | -14            |
| false         | 2   | 100 | 54             |

### Generate Last Name\{#generate-last-name}

The generate last name transformer generates a valid last name from a list of predefined last name values. You can see the entire list of last name value [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/last-names.json).

**Configurations**

This transformer has no configurations.

**Examples**

Here are some possible output examples:

| Example Output |
| -------------- |
| Dach           |
| Sporer         |
| Rodriguezon    |

### Generate SHA256 Hash\{#generate-sha256-hash}

The generate sha256 hash transformer generates a random SHA256 hash and hex encodes the resulting value and returns it as a string.

**Configurations**

There are no configurations for the hash transformer.

**Examples**

Here are some examples of what an output hash value may look like.

| Example Output                                                    |
| ----------------------------------------------------------------- |
| 1b836d86806e16bd45164b9d665ce86dac9d9d55bf79d0e771265da9fbf950d1  |
| c74f07e35232c844a6639fef0c82ad093351cfa1a1b458fd2afc44cc19211508C |
| 18dd7ce288e5538fbdd9c736aa9951ddb317e42c0b928e2b4db672414a82f811  |

### Generate SSN\{#generate-ssn}

The generate ssn transformer randomly generates a social security number and returns it with hyphens as a string.

For, example:
`123-45-6789`

**Configurations**

There are no configurations for the ssn transformer.

**Examples**

Here are some examples of what an output street address value may look like.

| Example Output |
| -------------- |
| 892-23-7634    |
| 479-82-3224    |

### Generate State\{#generate-state}

The generate state transformer generates a randomly selected US state. You can see the complete list of states that are available to be randomly selected [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/addresses.json).

If you'd rather not get back a real state value, you can use the [Random String Transformer](/transformers/system#generate-random-string) to generate a random string value.

**Configurations**

There are no configurations for the state transformer.

**Examples**

Here are some examples of what an output state value may look like.

| Example Output |
| -------------- |
| Rhode Island   |
| Missouri       |

### Generate Street Address\{#generate-street-address}

The generate street address transformer generates a randomly selects a real street address that exists in the United States. You can see the complete list of street addresses that are available to be randomly selected [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/addresses.json).

The street address transformer returns a valid United States address. For example:
`123 Main Street Boston`

**Configurations**

There are no configurations for the street address transformer.

**Examples**

Here are some examples of what an output street address value may look like.

| Example Output                |
| ----------------------------- |
| 509 Franklin Street Northeast |
| 14 Huntingon Street           |

### Generate String Phone Number\{#generate-string-phone-number}

The generate string phone transformer generates a random string phone number. By default, the string phone transformer generates a random 10 digit phone number with no hyphens.

**Configurations**

Here the configurations for the Generate String Phone Number Transformer:

| Name | Description                                                            | Default | Example Output |
| ---- | ---------------------------------------------------------------------- | ------- | -------------- |
| Min  | Min is the lower range of the a string length. The minimum range is 9. | 9       | 234232974865   |
| Max  | Max is the upper range of a length value. The maximum range is 15.     | 15      | 2037486752     |

**Examples**

There are several ways you can mix-and-match configurations to get different potential phone number formats. Here are some possible combinations:

| Min | Max | Example Output |
| --- | --- | -------------- |
| 10  | 14  | 5209273239     |
| 11  | 15  | 520927323912   |
| 12  | 12  | 520927323969   |

### Generate Random String\{#generate-random-string}

The generate random string transformer generates a random string of alphanumeric characters.

For example:
`h2983hf28h92`

By default, the random string transformer generates a string of 10 characters long.

**Configurations**

Depending on your validations, you may want to configure the output string. The generate string transformer has the following configurations:

| Name | Description                                                                         | Default | Example Output |
| ---- | ----------------------------------------------------------------------------------- | ------- | -------------- |
| Min  | Specifies the min range for a generated int64 value. This supports negative values. | 2       | jn9d           |
| Max  | Specifies the max range for a generated int64 value. This supports negative values. | 7       | hd9n93         |

**Examples**

There are several ways you can mix-and-match configurations to get different potential random integer formats. Here are some possible combinations:

| Min | Max | Example Output |
| --- | --- | -------------- |
| 2   | 7   | je7R6          |
| 1   | 20  | 29h23rega9     |
| 5   | 5   | h2dni          |

### Generate Unix Timestamp\{#generate-unix-timestamp}

The generate unix timestamp transformer randomly generates a unix timestamp in UTC timezone and returns back an int64 representation of that timestamp.

By default, the generated timestamp will always be in the **past**.

**Configurations**

There are no configurations for this transformer.

**Examples**

Here are some examples of what an output state value may look like.

| Example Output |
| -------------- |
| 2524608000     |
| 946684800      |

### Generate Username\{#generate-username}

The generate username transformer generates a random string in the format of `<first_initial><last_name>`. The last names are pulled from the last name transformer(/transformers/system#generate-last-name).

**Configurations**

There are no configurations for the state transformer.

**Examples**

Here are some examples of what an output state value may look like.

| Example Output |
| -------------- |
| jdoe           |
| lsmith         |

### Generate UTC Timestamp\{#generate-utc-timestamp}

The generate utc timestamp transformer randomly generates a utc timestamp in UTC timezone and returns back an time.Time representation of that timestamp.

By default, the generated timestamp will always be in the **past**.

**Configurations**

There are no configurations for this transformer.

**Examples**

Here are some examples of what an output state value may look like.

| Example Output                   |
| -------------------------------- |
| July 8, 2025, 00:00:00 UTC       |
| September 21, 1989, 00:00:00 UTC |

### Generate UUID\{#generate-uuid}

The generate UUID transformer generates a new UUID v4.

For example:
`6d871028b072442c9ad9e6e4e223adfa`

**Configurations**

Depending on your validations, you may want to configure the output uuid. The uuid transformer has the following configurations:

| Name            | Description                                                                                                                                                | Default | Example Output                       |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------------------------------ |
| Include hyphens | Includes hyphens in the output UUID. Note: For some databases, such as Postgres, it will automatically convert any UUID without hyphens to having hyphens. | false   | 6817d39c-ea4f-45ec-a81c-bc0801355576 |

**Examples**

Here are some example UUID values that the uuid transformer can generate:

| Include Hyphens | Example Output                       |
| --------------- | ------------------------------------ |
| false           | 53ef1f16cd0d455cb45c4cb6434ae807     |
| true            | ab6b676b-0d0e-4e38-b98a-3935a832da7d |

### Generate Zipcode\{#generate-zipcode}

The generate zipcode transformer generates a randomly selected US zipcode. You can see the complete list of zipcodes that are available to be randomly selected [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/addresses.json).

If you'd rather not get back a real zipcode value, you can use the [Random String Transformer](/transformers/system#generate-random-string) to generate a random string value.

**Configurations**

There are no configurations for the zipcode transformer.

**Examples**

Here are some examples of what an output zipcode value may look like.

| Example Output |
| -------------- |
| 27563          |
| 90237          |

### Transform E164 Phone Number\{#transform-e164-phone-number}

The transform e164 phone transformer can anonymize an existing e164 phone number or completely generate a new one. It returns a string value with the format `+<number>`.

For example, the following input value:
`+7829828714`

Could produce the following output value:
`+5698437232`

E164 Phone numbers vary in length with some international phone numbers reaching up to 15 digits in length. You can set this transformer to respect the length of the input value in order maintain the same shape as the input value.

**Configurations**

Depending on your validations, you may want to configure the output e164 phone number. The transform e164 phone number transformer has the following configurations:

| Name            | Description                                                                                                                                                                             | Default | Example Input | Example Output |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| Preserve Length | Preserve Length will ensure that the output e164 phone number is the same length as the input phone number. If set to false, it will generate a value between 9 and 15 characters long. | false   | +8923786243   | +290374867526  |

**Examples**

Here are some examples of what the transform e164 phone number transformer can output:

| Preserve Length | Example Input | Example Output |
| --------------- | ------------- | -------------- |
| false           | +2890923784   | +520927392     |
| true            | +62890923784  | +52092732393   |

### Transform First Name\{#transform-first-name}

The transform first name transformer generates a valid first name from a list of predefined first name values. You can see the entire list of first name value [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/first-names.json).

By default, the first name transformer generates a first name of random length. To preserve the length of the input first name, you can set the `preserveLength` config.

**Configurations**

Depending on your validations, you may want to configure the output first name. The first name transformer has the following configurations:

| Name           | Description                                                                                                                                                                 | Default | Example Input | Example Output |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length will ensure that the output first name is the same length as the input first name. The preserveLength config only preserves names up to 12 characters long. | false   | Johnathan     | Bill           |

**Examples**

Here are some examples of what the transform first name transformer can output:

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | Joe           | Benjamin       |
| true           | Frank         | Dante          |

### Transform Float64\{#transform-float64}

The transform float transformer can anonymize an existing float64 value or generate a completely random floating point number.

For example:
`32.2432`

The params `randomizationRangeMin` and `randomizationRangeMax` set an upper and lower bound around the value that you want to anonymize relative to the input. For example, if the input value is 10, and you set the `randomizationRangeMin` value to 5, then the minimum will be 5. And if you set the `randomizationRangeMax` to 5, then the maximum will be 15 ( 10 + 5 = 15).

**Configurations**

Depending on your validations, you may want to configure the output float. The random float transformer has the following configurations:

| Name               | Description                                          | Default |
| ------------------ | ---------------------------------------------------- | ------- |
| Relative Range Min | Sets the lower bound of a range for the input value. | 20.00   |
| Relative Range Max | Sets the upper bound of a range for the input value. | 50.00   |

**Examples**

There are several ways you can mix-and-match configurations to get different potential random float formats. Here are some possible combinations:

| Randomization Range Min | Randomization Range Max | Example Input | Example Output |
| ----------------------- | ----------------------- | ------------- | -------------- |
| 10                      | 20                      | 27.23         | 17.248         |
| 14                      | 29                      | 2.344         | -9.24          |
| 5                       | 10                      | -8.23         | -4.19          |

### Transform Full Name\{#transform-full-name}

The transform full name transformer generates a valid full name from a list of predefined full name values. The generated full name is made from a combination of the [first](/transformers/system#generate-first-name) and [last](/transformers/system#generate-last-name) names transformers.

By default, the full name transformer generates a full name of random length. To preserve the length of the input full name, you can set the `preserveLength` config.

**Configurations**

Depending on your validations, you may want to configure the output full name. The full name transformer has the following configurations:

| Name           | Description                                                                                                                                                                                                                                                                                    | Default | Example Input | Example Output |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length will ensure that the output full name is the same length as the input full name by preserving the length of the first and last names. The preserveLength config only preserves first and/or last names up to 12 characters long. So the max full name length is 25 characters. | false   | John          | Bill           |

**Examples**

There are several ways you can mix-and-match configurations to get different full name formats. Here are some possible combinations:

| PreserveLength | Example Input | Example Output  |
| -------------- | ------------- | --------------- |
| false          | Joe Smith     | Edward Marshall |
| true           | Frank Jones   | Dante Mayer     |
| false          | Aidan Bell    | Gregory Johnson |

### Transform Int64 Phone Number\{#transform-int64-phone-number}

The transform int64 phone transformer can anonymize an existing phone number or completely generate a new one. By default, the transform int64 phone number transformer generates a random 10 digit phone number.

For example, the following input value:
`7829828714`

Would produce the following output value:
`5698437232`

You can see we generated a new phone integer value that can be used as an integer phone number. Also, note that we don't include hyphens in this transformer since the output type is an `integer`.

**#Configurations**

Depending on your validations, you may want to configure the output phone number.

| Name           | Description                                                                                            | Default | Example Input | Example Output |
| -------------- | ------------------------------------------------------------------------------------------------------ | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length will ensure that the output phone number is the same length as the input phone number. | false   | 892387786243  | 2903748675392  |

**Examples**

Here are some possible example output values:

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | 2890923784    | 5209273239     |
| true           | 3839304957    | 6395745371     |

### Transform Int64\{#transform-int64}

The random integer transformer generates a random integer.

For example:
`6782`

**Configurations**

Depending on your validations, you may want to configure the output integer. The random integer transformer has the following configurations:

| Name               | Description                                                  | Default |
| ------------------ | ------------------------------------------------------------ | ------- |
| Relative Range Min | Sets the lower bound of a range relative to the input value. | 20      |
| Relative Range Max | Sets the upper bound of a range relative to the input value. | 50      |

**Examples**

There are several ways you can mix-and-match configurations to get different potential random int64 values. Here are some possible combinations:

| Relative Range Min | Relative Range Max | Example Input | Example Output |
| ------------------ | ------------------ | ------------- | -------------- |
| 5                  | 5                  | 10            | 15             |
| 10                 | 70                 | 23            | 91             |
| 20                 | 90                 | 30            | 27             |

### Transform Last Name\{#transform-last-name}

The transform last name transformer generates a valid last name from a list of predefined last name values. You can see the entire list of last name value [here.](https://github.com/nucleuscloud/neosync/blob/main/worker/internal/benthos/transformers/data-sets/last-names.json).

By default, the last name transformer generates a last name of random length. To preserve the length of the input last name, you can set the `preserveLength` config.

**Configurations**

Depending on your validations, you may want to configure the output last name. The last name transformer has the following configurations:

| Name           | Description                                                                                                                                                               | Default | Example Input | Example Output |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length will ensure that the output last name is the same length as the input last name. The preserveLength config only preserves names up to 12 characters long. | false   | Hills         | Kunze          |

**Examples**

There are several ways you can mix-and-match configurations to get different last name formats. Here are some possible combinations:

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | Lei           | Dach           |
| true           | Littel        | Sporer         |
| false          | Johnsonston   | Rodriguezon    |

### Transform Phone Number\{#transform-phone-number}

The transform phone transformer can anonymize an existing phone number or completely generate a new one. This transformer specifically takes a string value and returns a string value.

By default, the transform phone transformer generates a random 10 string value phone number with no hyphens.

For example, the following input value:
`7829828714`

Would produce the following output value:
`5698437232`

**Configurations**

Depending on your validations, you may want to configure the output phone number. Here are the configurations for the transform string phone number transformer.

| Name           | Description                                                                                            | Default | Example Input   | Example Output  |
| -------------- | ------------------------------------------------------------------------------------------------------ | ------- | --------------- | --------------- |
| PreserveLength | Preserve Length will ensure that the output phone number is the same length as the input phone number. | false   | 892387243786243 | 290374867526392 |

**Examples**

There are several ways you can mix-and-match configurations to get different potential phone number formats. Here are some possible combinations:

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | 2890923784    | 520927323239   |
| true           | 2890923784    | 5209223539     |

### Transform String\{#transform-string}

The random string transformer generates a random string of alphanumeric characters.

For example:
`h2983hf28h92`

By default, the random string transformer generates a string of 10 characters long.

**Configurations**

Depending on your validations, you may want to configure the output string. The random string transformer has the following configurations:

| Name            | Description                                                     | Default | Example Input | Example Output |
| --------------- | --------------------------------------------------------------- | ------- | ------------- | -------------- |
| Preserve Length | Preserves the length of the input integer to the output string. | false   | hello         | 9Fau3          |

**Examples**

There are several ways you can mix-and-match configurations to get different potential random integer formats. Here are some possible combinations:

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | bill          | je7R6          |
| true           | lorem         | i1NMe          |

### Transform Character Scramble\{#transform-character-scramble}

Transforms a string value with characters into an anonymized version of that string value while preserving spaces and capitalization. Letters will be replaced with letters, numbers with numbers and non-number or letter ASCII characters such as !&\* with other characters.

For example:

Original: Hello World 123!$%
Substituted: Ifmmp Xpsme 234@%^

Note that this does not work for hex values: 0x00 -> 0x1F such as chinese characters.

**Configurations**

| Name  | Description                                                                                                                       | Default | Example Input | Example Output |
| ----- | --------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| Regex | A regular expression to match a part of the string to scramble. If no regex is specified that we will scramble the entire string. | empty   | ello          | hello -> hmato |

**Examples**

There are several ways you can mix-and-match configurations to get different potential random integer formats. Here are some possible combinations:

| Regex    | Example Input | Example Output |
| -------- | ------------- | -------------- |
|          | helloworld    | didlenfioq     |
| ello     | Hello World!  | Hjiqs World!   |
| Hello 12 | Hello 1234    | Praji 2834     |

### Passthrough\{#passthrough}

The passthrough transformer simplify passes the input data out to the output without making any modifications to it. This is useful in many circumstances but cautious of accidentally leaking sensitive data through this transformer.

### Null\{#null}

The null transformer simply returns a null value. This may be useful if you a column that can't be null but don't have a specific value that you want to insert.

**Configurations**

There are no configurations for the null transformer.

**Examples**

Here are some examples of what an output null value may look like.

| Example Input | Example Output |
| ------------- | -------------- |
| N/A           | null           |

## Works-in-progress

This section is for detailing what we have limited or in-progress support for in regards to data transformations.

### Array data types

Neosync has limited support for array data types today.

1. Passthrough - Array data types can be set to passthrough and have their values directly passed through from the source to the destination.
2. Null - If the column is nullable, this transformer can be safely set.
3. Default - If the column has a default set, this transformer can be set to allow the column be set to its default state.
4. Javascript Code - The JS code transformer has support for array data types as the logic is effectively up to the function's implementation.
