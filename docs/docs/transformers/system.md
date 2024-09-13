---
title: System Transformers
description: Learn about Neosync's System Transformers that come out of the box and help anonymize data and generate synthetic data
id: system
hide_title: false
slug: /transformers/system
# cSpell:words luhn,huntingon,dach,sporer,rodriguezon,rega,jdoe,lsmith,aidan,kunze,littel,johnsonston
---

## Introduction

Neosync ships with 40+ System Transformers to give you an easy way to get started with anonymizing or generating data. There are two kinds of System Transformers:

1. Generators - these begin with `Generate` and do not anonymize existing data. They only generate net-new synthetic data.
2. Transformers - these begin with `Transform` and can anonymize existing data or generate new-new synthetic data based on your configuration.

## Reference

| Name                                                                              | Type    | Description                                                                                                                      |
| --------------------------------------------------------------------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------- |
| [Generate Categorical](/transformers/system#generate-categorical)                 | string  | Randomly selects a value from a defined set of categorical values                                                                |
| [Generate Email](/transformers/system#generate-email)                             | string  | Generates a new randomized email address.                                                                                        |
| [Generate Boolean](/transformers/system#generate-bool)                            | boolean | Generates a boolean value at random.                                                                                             |
| [Generate Card Number](/transformers/system#generate-card-number)                 | int64   | Generates a card number.                                                                                                         |
| [Generate City](/transformers/system#generate-city)                               | string  | Randomly selects a city from a list of predefined US cities.                                                                     |
| [Generate Country](/transformers/system#generate-country)                         | string  | Randomly selects a country and returns it as a 2-letter country code or the full country name                                    |
| [Generate E164 Phone Number](/transformers/system#generate-e164-phone-number)     | string  | Generates a Generate phone number in e164 format.                                                                                |
| [Generate First Name](/transformers/system#generate-first-name)                   | string  | Generates a random first name.                                                                                                   |
| [Generate Float64](/transformers/system#generate-float64)                         | float64 | Generates a random float64 value.                                                                                                |
| [Generate Full Address](/transformers/system#generate-full-address)               | string  | Randomly generates a street address.                                                                                             |
| [Generate Full Name](/transformers/system#generate-full-name)                     | string  | Generates a new full name consisting of a first and last name.                                                                   |
| [Generate Gender](/transformers/system#generate-gender)                           | string  | Randomly generates one of the following genders: female, male, undefined, nonbinary.                                             |
| [Generate Javascript](/transformers/system#generate-javascript)                   | any     | Executes provided javascript code in the transformer for every row                                                               |
| [Generate Int64 Phone Number](/transformers/system#generate-int64-phone-number)   | int64   | Generates a new phone number of type int64 with a default length of 10.                                                          |
| [Generate Random Int64](/transformers/system#generate-int64)                      | int64   | Generates a random integer value with a default length of 4 unless the Integer Length or Preserve Length parameters are defined. |
| [Generate Last Name](/transformers/system#generate-last-name)                     | int64   | Generates a random last name.                                                                                                    |
| [Generate SHA256 Hash](/transformers/system#generate-sha256hash)                  | string  | SHA256 hashes a randomly generated value.                                                                                        |
| [Generate SSN](/transformers/system#generate-ssn)                                 | string  | Generates a completely random social security numbers including the hyphens in the format `xxx-xx-xxxx`.                         |
| [Generate State](/transformers/system#generate-state)                             | string  | Randomly selects a US state and returns the two-character state code.                                                            |
| [Generate Street Address](/transformers/system#generate-street-address)           | string  | Randomly generates a street address.                                                                                             |
| [Generate String Phone Number](/transformers/system#generate-string-phone-number) | string  | Generates a Generate phone number and returns it as a string.                                                                    |
| [Generate Random String](/transformers/system#generate-random-string)             | string  | Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length parameter are defined.       |
| [Generate Unix Timestamp](/transformers/system#generate-unixtimestamp)            | int64   | Randomly generates a Unix timestamp.                                                                                             |
| [Generate Username](/transformers/system#generate-username)                       | string  | Randomly generates a username                                                                                                    |
| [Generate UTC Timestamp](/transformers/system#generate-utctimestamp)              | time    | Randomly generates a UTC timestamp.                                                                                              |
| [Generate UUID](/transformers/system#generate-uuid)                               | uuid    | Generates a new UUIDv4 id.                                                                                                       |
| [Generate Zip code](/transformers/system#generate-zipcode)                        | string  | Randomly selects a zip code from a list of predefined US zip codes.                                                              |
| [Transform Email](/transformers/system#transform-email)                           | string  | Transforms an existing email address.                                                                                            |
| [Transform E164 Phone Number](/transformers/system#transform-e164-phone-number)   | string  | Transforms an existing E164 formatted phone number.                                                                              |
| [Transform First Name](/transformers/system#transform-first-name)                 | string  | Transforms an existing first name                                                                                                |
| [Transform Float64](/transformers/system#transform-float64)                       | float64 | Transforms an existing float value.                                                                                              |
| [Transform Full Name](/transformers/system#transform-full-name)                   | string  | Transforms an existing full name.                                                                                                |
| [Transform Int64 Phone Number](/transformers/system#transform-int64-phone-number) | int64   | Transforms an existing phone number that is typed as an integer                                                                  |
| [Transform Int64](/transformers/system#transform-int64)                           | int64   | Transforms an existing integer value.                                                                                            |
| [Transform Javascript](/transformers/system#transform-javascript)                 | any     | Executes user-provided javascript in on the input value.                                                                         |
| [Transform Last Name](/transformers/system#transform-last-name)                   | string  | Transforms an existing last name.                                                                                                |
| [Transform Phone Number](/transformers/system#transform-phone-number)             | string  | Transforms an existing phone number that is typed as a string.                                                                   |
| [Transform String](/transformers/system#transform-string)                         | string  | Transforms an existing string value.                                                                                             |
| [Transform Character Scramble](/transformers/system#transform-character-scramble) | string  | Transforms an existing string value by scrambling the characters while maintaining the format.                                   |
| [Passthrough](/transformers/system#passthrough)                                   | string  | Passes the input value through to the destination with no changes.                                                               |
| [Null](/transformers/system#generate-null)                                        | string  | Inserts a `null` string instead of the source value.                                                                             |
| [Use Column Default](/transformers/system#generate-default)                       | any     | Applies the predefined default value of a column as specified in the SQL database schema                                         |

### Generate Categorical\{#generate-categorical}

Allows a user to define a list of comma-separated string values that Neosync will randomly sample.

**Configurations**

| Name       | Description                                                              | Default         | Example Output |
| ---------- | ------------------------------------------------------------------------ | --------------- | -------------- |
| Categories | List of comma-separated string values that Neosync will randomly sample. | `value1,value2` | `value1`       |

**Examples**

| Categories | Example Input    | Example Output |
| ---------- | ---------------- | -------------- |
| false      | `red,blue,green` | `green`        |
| false      | `red,blue,green` | `red`          |

### Generate Email\{#generate-email}

Generates a new randomized email address in the format:

`<username>@<domain>.<top-level-domain>`

By default, the transformer randomizes the username, domain and top-level domain while always preserving the email format by retaining the `@` and `.` characters.

**Configurations**

| Name       | Description                                                                         | Default | Example Output                                 |
| ---------- | ----------------------------------------------------------------------------------- | ------- | ---------------------------------------------- |
| Email Type | Provides a way to generate unique email addresses by appending a UUID to the domain | uuid_v4 | ab6b676b-0d0e-4e38-b98a-3935a832da7d@gmail.com |

**Examples**

| Example Output                         |
| -------------------------------------- |
| `ab6b676b-0d0e-4e38-b98a-3935a832da7d` |
| `jFrankd@msn.com`                      |

### Generate Boolean\{#generate-bool}

Randomly generates a boolean value.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| true           |
| false          |

### Generate Card Number\{#generate-card-number}

Generates a new card number that can be [luhn valid.](https://en.wikipedia.org/wiki/Luhn_algorithm)

By default, generates a random 16 digit card number that is _not_ luhn valid. If you want luhn validation, please set the luhn-check config to `true`.

**Configurations**

| Name       | Description                                           | Default | Example Output   |
| ---------- | ----------------------------------------------------- | ------- | ---------------- |
| Valid Luhn | Generate a card number that can pass luhn validation. | true    | 6621238134011099 |

**Examples**

| Valid Luhn | Example Output   |
| ---------- | ---------------- |
| false      | 3627893345931223 |
| true       | 6621238134011099 |

### Generate City\{#generate-city}

Generates a randomly selected US city. You can see the complete list of cities that are available to be randomly selected

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| Louisville     |
| Manchester     |

### Generate Country\{#generate-country}

Generates a randomly selected country. It can return it in either a 2-letter country code format or the full country name.

**Configurations**

| Name      | Description                                                | Default | Example Output |
| --------- | ---------------------------------------------------------- | ------- | -------------- |
| Full Name | If true, the transformer will return the full country name | false   | AL             |

**Examples**

| Example Output |
| -------------- |
| Canada         |
| AL             |

### Generate E164 Phone Number\{#generate-e164-phone-number}

Generates a new international phone number including the + sign. By default, the generate e164 transformer generates a random phone number with no hyphens in the format:

`+34567890123`

Phone numbers also vary in length with some international phone numbers reaching up to 15 digits in length. You can set the range of the value by setting the `min` and `max` params. Here is more information on the [e164 format](https://www.twilio.com/docs/glossary/what-e164). This transformer also has a min of 9 and a max of 15. If you want to generate a number that is longer than that, you can use the [Generate Random int64 transformer](/transformers/system#generate-random-int64)

**Configurations**

| Name | Description                                   | Default | Example Output |
| ---- | --------------------------------------------- | ------- | -------------- |
| Min  | Min is the lower range of the a length value. | 9       | +29748675      |
| Max  | Max is the upper range of a length value.     | 12      | +2037486752    |

**Examples**

| Min | Max | Example Output |
| --- | --- | -------------- |
| 9   | 12  | +5209273239    |
| 10  | 14  | +5209273239    |
| 12  | 12  | +52092732323   |

### Generate First Name\{#generate-first-name}

Generates a valid first name from a list of predefined first name values.

By default, the generate first name transformer randomly picks a first name with a length between 2 and 12.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| Benjamin       |
| Gregory        |
| Ashley         |

### Generate Float64\{#generate-float64}

Generates a random floating point number.

For example:
`32.2432`

This transformer supports both positive and negative float numbers with a max precision of 17.

Go float64 adheres to the IEEE 754 standard for double-precision floating-point numbers.

If this is not sufficient, contact us about creating a generator that supports numbers with higher precision.

**Configurations**

| Name          | Description                                                                                           | Default |
| ------------- | ----------------------------------------------------------------------------------------------------- | ------- |
| RandomizeSign | Randomly sets the sign of the value.                                                                  | false   |
| Min           | Specifies the min range for a generated float value.                                                  | 1.00    |
| Max           | Specifies the max range for a generated float value.                                                  | 100.00  |
| Precision     | Sets the precision of the value. For example, a precision of 6 will generate a value such as 42.9812. | 6       |

**Examples**

Note: if randomize sign has been selected, this may cause the generated numbers to exist outside of the configured min/max range.

| RandomizeSign | Min    | Max   | Precision | Example Output |
| ------------- | ------ | ----- | --------- | -------------- |
| true          | 2.00   | 35.00 | 4         | -23.84         |
| true          | 4.12   | 19.43 | 3         | 7.28           |
| false         | -30.00 | 2.00  | 8         | -15.54219      |
| false         | -20.00 | 30.00 | 4         | 20.19          |

### Generate Full Address\{#generate-full-address}

Generates a randomly selected real full address that exists in the United States.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output                                      |
| --------------------------------------------------- |
| 509 Franklin Street Northeast Washington, DC, 20017 |
| 14 Huntingon Street Manchester, CT 06040            |

### Generate Full Name\{#generate-full-name}

Generates a valid full name from a list of predefined full name values.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output  |
| --------------- |
| Edward Marshall |
| Dante Mayer     |
| Gregory Johnson |

### Generate Gender\{#generate-gender}

Randomly selects a gender value from a predefined list of genders. Here is the list:

| Gender    | Abbreviation |
| --------- | ------------ |
| male      | m            |
| female    | f            |
| nonbinary | n            |
| undefined | u            |

By default, the gender transformer does not abbreviate the gender. If you'd like to return an abbreviated gender, pass in the `abbreviate` config.

**Configurations**

| Name       | Description                                                                    | Default | Example Output |
| ---------- | ------------------------------------------------------------------------------ | ------- | -------------- |
| Abbreviate | Abbreviate will abbreviate the output gender so that it is only one character. | false   | u              |

**Examples**

| Abbreviate | Example Output |
| ---------- | -------------- |
| false      | male           |
| true       | f              |
| false      | nonbinary      |

### Generate Int64 Phone Number\{#generate-int64-phone-number}

Generates a random 10 digit phone number with a valid US area code and returns it as an `int64` type with no hyphens. If you want to return a `string` or want to include hyphens, check out the [Generate String Phone transformer](/transformers/system#generate-string-phone-number).

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| 5209273239     |
| 9378463528     |

### Generate Javascript\{#generate-javascript}

Allows the user to define Javascript code and then execute that on every row that they are generating. This transformer supports JS ECMA 5 and most of the standard JS lib. It does not currently support importing third party libraries.

**Configurations**

| Name | Description                               | Default         | Example Output |
| ---- | ----------------------------------------- | --------------- | -------------- |
| Code | The javascript code that will be executed | `return 'test'` | `test`         |

**Example**

| Code            | Example Output |
| --------------- | -------------- |
| `return 'test'` | `test`         |

### Generate Random Int64\{#generate-int64}

Generates a random integer and returns it as a int64 type.

For example:
`6782`

**Configurations**

| Name          | Description                                          | Default |
| ------------- | ---------------------------------------------------- | ------- |
| RandomizeSign | Randomly sets the sign of the generated value.       | false   |
| Min           | Specifies the min range for a generated int64 value. | 1       |
| Max           | Specifies the max range for a generated int64 value. | 40      |

**Examples**

Note: if randomize sign has been selected, this may cause the generated numbers to exist outside of the configured min/max range.

| RandomizeSign | Min | Max | Example Output |
| ------------- | --- | --- | -------------- |
| true          | 10  | 40  | -23            |
| true          | -20 | -3  | -14            |
| false         | 2   | 100 | 54             |

### Generate Last Name\{#generate-last-name}

Generates a valid last name from a list of predefined last name values.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| Dach           |
| Sporer         |
| Rodriguezon    |

### Generate SHA256 Hash\{#generate-sha256hash}

Generates a random SHA256 hash and hex encodes the resulting value and returns it as a string.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output                                                    |
| ----------------------------------------------------------------- |
| 1b836d86806e16bd45164b9d665ce86dac9d9d55bf79d0e771265da9fbf950d1  |
| c74f07e35232c844a6639fef0c82ad093351cfa1a1b458fd2afc44cc19211508C |
| 18dd7ce288e5538fbdd9c736aa9951ddb317e42c0b928e2b4db672414a82f811  |

### Generate SSN\{#generate-ssn}

Randomly generates a social security number and returns it with hyphens as a string.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| 892-23-7634    |
| 479-82-3224    |

### Generate State\{#generate-state}

Generates a randomly selected US state in either 2-letter state code format or the full state name.

**Configurations**

| Name               | Description                                                                                                            | Default | Example Output |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------- | ------- | -------------- |
| Generate Full Name | Set to true to return the full state name with a capitalized first letter. Returns the 2-letter state code by default. | false   | CA             |

**Examples**

| State Code | Example Output |
| ---------- | -------------- |
| False      | CA             |
| True       | California     |

### Generate Street Address\{#generate-street-address}

Generates a randomly selects a real street address that exists in the United States.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output                |
| ----------------------------- |
| 509 Franklin Street Northeast |
| 14 Huntingon Street           |

### Generate String Phone Number\{#generate-string-phone-number}

Generates a random string phone number. By default, the string phone transformer generates a random 10 digit phone number with no hyphens.

**Configurations**

| Name | Description                                                            | Default | Example Output |
| ---- | ---------------------------------------------------------------------- | ------- | -------------- |
| Min  | Min is the lower range of the a string length. The minimum range is 9. | 9       | 234232974865   |
| Max  | Max is the upper range of a length value. The maximum range is 15.     | 15      | 2037486752     |

**Examples**

| Min | Max | Example Output |
| --- | --- | -------------- |
| 10  | 14  | 5209273239     |
| 11  | 15  | 520927323912   |
| 12  | 12  | 520927323969   |

### Generate Random String\{#generate-random-string}

Generates a random string of alphanumeric characters.

For example:
`h2983hf28h92`

By default, the random string transformer generates a string of 10 characters long.

**Configurations**

| Name | Description                                                                         | Default | Example Output |
| ---- | ----------------------------------------------------------------------------------- | ------- | -------------- |
| Min  | Specifies the min range for a generated int64 value. This supports negative values. | 2       | jn9d           |
| Max  | Specifies the max range for a generated int64 value. This supports negative values. | 7       | hd9n93         |

**Examples**

| Min | Max | Example Output |
| --- | --- | -------------- |
| 2   | 7   | je7R6          |
| 1   | 20  | 29h23rega9     |
| 5   | 5   | h2dni          |

### Generate Unix Timestamp\{#generate-unixtimestamp}

Randomly generates a unix timestamp in UTC timezone and returns back an int64 representation of that timestamp.

By default, the generated timestamp will always be in the **past**.

**Configurations**

There are no configurations for this transformer.

**Examples**=

| Example Output |
| -------------- |
| 2524608000     |
| 946684800      |

### Generate Username\{#generate-username}

The generate username transformer generates a random string in the format of `<first_initial><last_name>`.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| jdoe           |
| lsmith         |

### Generate UTC Timestamp\{#generate-utctimestamp}

Randomly generates a utc timestamp in UTC timezone and returns back a time.Time representation of that timestamp.

By default, the generated timestamp will always be in the **past**.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output                   |
| -------------------------------- |
| July 8, 2025, 00:00:00 UTC       |
| September 21, 1989, 00:00:00 UTC |

### Generate UUID\{#generate-uuid}

Generates a new UUID v4.

For example:
`6d871028b072442c9ad9e6e4e223adfa`

**Configurations**

| Name            | Description                                                                                                                                                | Default | Example Output                       |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------------------------------ |
| Include hyphens | Includes hyphens in the output UUID. Note: For some databases, such as Postgres, it will automatically convert any UUID without hyphens to having hyphens. | false   | 6817d39c-ea4f-45ec-a81c-bc0801355576 |

**Examples**

| Include Hyphens | Example Output                       |
| --------------- | ------------------------------------ |
| false           | 53ef1f16cd0d455cb45c4cb6434ae807     |
| true            | ab6b676b-0d0e-4e38-b98a-3935a832da7d |

### Generate Zipcode\{#generate-zipcode}

Generates a randomly selected US zipcode.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| 27563          |
| 90237          |

### Transform Email\{#transform-email}

Anonymizes an existing email address or generate a new randomized email address in the format:

`<username>@<domain>.<top-level-domain>`

By default, the transformer randomizes the username, domain and top-level domain while always preserving the email format by retaining the `@` and `.` characters.

**Configurations**

| Name                 | Description                                                                                       | Default | Example Input   | Example Output                                 |
| -------------------- | ------------------------------------------------------------------------------------------------- | ------- | --------------- | ---------------------------------------------- |
| Preserve Length      | Preserve Length will ensure that the output email is the same length as the input                 | false   | john@gmail      | william@gmail.com                              |
| Preserve Domain      | Preserve Domain will ensure that the output email domain is the same as the input email domain    | false   | john@gmail      | william@yahoo.com                              |
| Excluded Domains     | Takes in a list of comma separated domains and excludes them from the transformer                 | []      | @gmail          | william@gmail.com                              |
| Email Type           | Provides a way to generate unique email addresses by appending a UUID to the domain               | uuid_v4 | john@gmail      | ab6b676b-0d0e-4e38-b98a-3935a832da7d@gmail.com |
| Invalid Email Action | Provide a way for Neosync to handle email strings that are not formatted as valid email addresses | reject  | john@gmail..com | null                                           |

**Examples**

There are several ways you can mix-and-match configurations to get different potential email formats. Note that the Excluded Domains List pertains to the transformed emails. If the PreserverDomain is true and there is a domain in the Excluded Domains list, then the default logic would be to **not** preserve the domain but since we are excluding those domains from the transformation logic via the exclude list, then those domains **will** be excluded.

| PreserveLength | PreserveDomain | Exclude Domains | Example Input        | Invalid Email Action | Example Output        |
| -------------- | -------------- | --------------- | -------------------- | -------------------- | --------------------- |
| false          | true           | `gmail.com`     | `evis@gmail.com`     | reject               | `f9nkeuh@ergwer.wewe` |
| false          | true           | `gmail.com`     | `evis@gmail....com`  | reject               | ``                    |
| true           | false          | `gmail.com`     | `f98723uh@gmail.com` | null                 | `f98723uh@gmail.com`  |
| true           | true           | `gmail.com`     | `evis@gmail.com`     | passthrough          | `f98723uh@weefw.wefw` |
| false          | false          | `gmail.com`     | `evis@gmail.com`     | passthrough          | `f98723uh@gmail.com`  |

### Transform E164 Phone Number\{#transform-e164-phone-number}

Anonymizes an existing e164 phone number or completely generate a new one. It returns a string value with the format `+<number>`.

E164 Phone numbers vary in length with some international phone numbers reaching up to 15 digits in length. You can set this transformer to respect the length of the input value in order maintain the same shape as the input value.

**Configurations**

| Name            | Description                                                                                                                                                                             | Default | Example Input | Example Output |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| Preserve Length | Preserve Length will ensure that the output e164 phone number is the same length as the input phone number. If set to false, it will generate a value between 9 and 15 characters long. | false   | +8923786243   | +290374867526  |

**Examples**

| Preserve Length | Example Input | Example Output |
| --------------- | ------------- | -------------- |
| false           | +2890923784   | +520927392     |
| true            | +62890923784  | +52092732393   |

### Transform First Name\{#transform-first-name}

Generates a valid first name from a list of predefined first name values.

By default, the first name transformer generates a first name of random length. To preserve the length of the input first name, you can set the `preserveLength` config.

**Configurations**

| Name           | Description                                                                                                                                                                 | Default | Example Input | Example Output |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length will ensure that the output first name is the same length as the input first name. The preserveLength config only preserves names up to 12 characters long. | false   | Johnathan     | Bill           |

**Examples**

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | Joe           | Benjamin       |
| true           | Frank         | Dante          |

### Transform Float64\{#transform-float64}

Anonymizes an existing float64 value or generate a completely random floating point number.

The params `randomizationRangeMin` and `randomizationRangeMax` set an upper and lower bound around the value that you want to anonymize relative to the input. For example, if the input value is 10, and you set the `randomizationRangeMin` value to 5, then the minimum will be 5. And if you set the `randomizationRangeMax` to 5, then the maximum will be 15 ( 10 + 5 = 15).

**Configurations**

| Name               | Description                                          | Default |
| ------------------ | ---------------------------------------------------- | ------- |
| Relative Range Min | Sets the lower bound of a range for the input value. | 20.00   |
| Relative Range Max | Sets the upper bound of a range for the input value. | 50.00   |

**Examples**

| Randomization Range Min | Randomization Range Max | Example Input | Example Output |
| ----------------------- | ----------------------- | ------------- | -------------- |
| 10                      | 20                      | 27.23         | 17.248         |
| 14                      | 29                      | 2.344         | -9.24          |
| 5                       | 10                      | -8.23         | -4.19          |

### Transform Full Name\{#transform-full-name}

Generates a valid full name from a list of predefined full name values.

By default, the full name transformer generates a full name of random length. To preserve the length of the input full name, you can set the `preserveLength` config.

**Configurations**

| Name           | Description                                                                                                                                                                                                                                                         | Default | Example Input | Example Output |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length ensures that the output full name is the same length as the input full name by preserving the length of the first and last names. It only preserves first and/or last names up to 12 characters long. So the max full name length is 25 characters. | false   | John          | Bill           |

**Examples**

| PreserveLength | Example Input | Example Output  |
| -------------- | ------------- | --------------- |
| false          | Joe Smith     | Edward Marshall |
| true           | Frank Jones   | Dante Mayer     |
| false          | Aidan Bell    | Gregory Johnson |

### Transform Int64 Phone Number\{#transform-int64-phone-number}

Anonymizes an existing phone number or completely generate a new one. By default, the transform int64 phone number transformer generates a random 10 digit phone number.

You can see we generated a new phone integer value that can be used as an integer phone number. Also, note that we don't include hyphens in this transformer since the output type is an `integer`.

**#Configurations**

| Name           | Description                                                                                            | Default | Example Input | Example Output |
| -------------- | ------------------------------------------------------------------------------------------------------ | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length will ensure that the output phone number is the same length as the input phone number. | false   | 892387786243  | 2903748675392  |

**Examples**

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | 2890923784    | 5209273239     |
| true           | 3839304957    | 6395745371     |

### Transform Int64\{#transform-int64}

Anonymizes an existing int64 value.

**Configurations**

| Name               | Description                                                  | Default |
| ------------------ | ------------------------------------------------------------ | ------- |
| Relative Range Min | Sets the lower bound of a range relative to the input value. | 20      |
| Relative Range Max | Sets the upper bound of a range relative to the input value. | 50      |

**Examples**

| Relative Range Min | Relative Range Max | Example Input | Example Output |
| ------------------ | ------------------ | ------------- | -------------- |
| 5                  | 5                  | 10            | 15             |
| 10                 | 70                 | 23            | 91             |
| 20                 | 90                 | 30            | 27             |

### Transform Javascript\{#transform-javascript}

Executes user provided code on the incoming input value. It's a very flexible way to transform the incoming input value however you'd like.

For example:

```js
function append(value) {
  return value + 'this_is_transformed';
}

return append(value);
```

**Configurations**

| Name | Description            | Default               |
| ---- | ---------------------- | --------------------- |
| Code | The user provided code | `return value + test` |

**Examples**

| Relative Range Min | Relative Range Max | Example Input | Example Output |
| ------------------ | ------------------ | ------------- | -------------- |
| 5                  | 5                  | 10            | 15             |
| 10                 | 70                 | 23            | 91             |
| 20                 | 90                 | 30            | 27             |

### Transform Last Name\{#transform-last-name}

Generates a valid last name from a list of predefined last name values.

By default, the last name transformer generates a last name of random length. To preserve the length of the input last name, you can set the `preserveLength` config.

**Configurations**

| Name           | Description                                                                                                                                                               | Default | Example Input | Example Output |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- | -------------- |
| PreserveLength | Preserve Length will ensure that the output last name is the same length as the input last name. The preserveLength config only preserves names up to 12 characters long. | false   | Hills         | Kunze          |

**Examples**

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | Lei           | Dach           |
| true           | Littel        | Sporer         |
| false          | Johnsonston   | Rodriguezon    |

### Transform Phone Number\{#transform-phone-number}

Anonymizes an existing phone number or completely generate a new one. This transformer specifically takes a string value and returns a string value.

By default, the transform phone transformer generates a random 10 string value phone number with no hyphens.

**Configurations**

| Name           | Description                                                                                            | Default | Example Input   | Example Output  |
| -------------- | ------------------------------------------------------------------------------------------------------ | ------- | --------------- | --------------- |
| PreserveLength | Preserve Length will ensure that the output phone number is the same length as the input phone number. | false   | 892387243786243 | 290374867526392 |

**Examples**

| PreserveLength | Example Input | Example Output |
| -------------- | ------------- | -------------- |
| false          | 2890923784    | 520927323239   |
| true           | 2890923784    | 5209223539     |

### Transform String\{#transform-string}

Anonymizes an existing string value of alphanumeric characters.

By default, the random string transformer generates a string of 10 characters long.

**Configurations**

| Name            | Description                                                     | Default | Example Input | Example Output |
| --------------- | --------------------------------------------------------------- | ------- | ------------- | -------------- |
| Preserve Length | Preserves the length of the input integer to the output string. | false   | hello         | 9Fau3          |

**Examples**

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

| Regex    | Example Input | Example Output |
| -------- | ------------- | -------------- |
|          | helloworld    | didlenfioq     |
| ello     | Hello World!  | Hjiqs World!   |
| Hello 12 | Hello 1234    | Praji 2834     |

### Passthrough\{#passthrough}

Passes the input data out to the output without making any modifications to it. This is useful in many circumstances but cautious of accidentally leaking sensitive data through this transformer.

### Null\{#generate-null}

Simply returns a null value. This may be useful if you a column that can't be null but don't have a specific value that you want to insert.

**Configurations**

There are no configurations for this transformer.

**Examples**

Here are some examples of what an output null value may look like.

| Example Input | Example Output |
| ------------- | -------------- |
| N/A           | null           |

### Use Column Default\{#generate-default}

Automatically applies the predefined default value of a column as specified in the SQL database schema. This means whenever new data is added without specifying a value for this column, the system will insert the default value set for that column in the database. This is typically used in conjunction with `GENERATED` columns.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Input | Example Output |
| ------------- | -------------- |
| 1             | 1              |

## Works-in-progress

This section is for detailing what we have limited or in-progress support for in regards to data transformations.

### Array data types

Neosync has limited support for array data types today.

1. Passthrough - Array data types can be set to passthrough and have their values directly passed through from the source to the destination.
2. Null - If the column is nullable, this transformer can be safely set.
3. Default - If the column has a default set, this transformer can be set to allow the column be set to its default state.
4. Javascript Code - The JS code transformer has support for array data types as the logic is effectively up to the function's implementation.
