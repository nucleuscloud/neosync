---
title: System Transformers
description: Learn about Neosync's System Transformers that come out of the box and help anonymize data and generate synthetic data
id: system
hide_title: false
slug: /transformers/system
# cSpell:words luhn,huntingon,dach,sporer,rodriguezon,rega,jdoe,lsmith,aidan,kunze,littel,johnsonston,unixtimestamp,utctimestamp
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

Randomly selects a value from a defined set of categorical values.

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

Generates a new randomized email address.

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

Generates a random boolean value.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| true           |
| false          |

### Generate Card Number\{#generate-card-number}

Generates a 16 digit card number that is valid by Luhn valid by default.

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

Randomly selects a city from a list of predefined US cities.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| Louisville     |
| Manchester     |

### Generate Country\{#generate-country}

Randomly selects a country and by default, returns it as a 2-letter country code.

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

Generates a new random international phone number including the + sign and no hyphens.

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

Generates a random first name between 2 and 12 characters long.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| Benjamin       |
| Gregory        |
| Ashley         |

### Generate Float64\{#generate-float64}

Generates a random floating point number with a max precision of 17. Go float64 adheres to the IEEE 754 standard for double-precision floating-point numbers.

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

Generates a new full name consisting of a first and last name.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output  |
| --------------- |
| Edward Marshall |
| Dante Mayer     |
| Gregory Johnson |

### Generate Gender\{#generate-gender}

Randomly generates one of the following genders: female (f), male (m), undefined (u), nonbinary (n).

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

Generates a new int64 phone number with a default length of 10.

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

Generates a random int64 value with a default length of 4.

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

Generates a random last name.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| Dach           |
| Sporer         |
| Rodriguezon    |

### Generate SHA256 Hash\{#generate-sha256hash}

Generates a random SHA256 hash and returns it as a string.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output                                                    |
| ----------------------------------------------------------------- |
| 1b836d86806e16bd45164b9d665ce86dac9d9d55bf79d0e771265da9fbf950d1  |
| c74f07e35232c844a6639fef0c82ad093351cfa1a1b458fd2afc44cc19211508C |
| 18dd7ce288e5538fbdd9c736aa9951ddb317e42c0b928e2b4db672414a82f811  |

### Generate SSN\{#generate-ssn}

Generates a random social security numbers including the hyphens in the format xxx-xx-xxxx.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| 892-23-7634    |
| 479-82-3224    |

### Generate State\{#generate-state}

Randomly selects a US state and by default, returns it as a 2-letter state code.

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

Randomly generates a street address.

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output                |
| ----------------------------- |
| 509 Franklin Street Northeast |
| 14 Huntingon Street           |

### Generate String Phone Number\{#generate-string-phone-number}

Generates a random 10 digit phone number and returns it as a string with no hyphens.

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

Randomly generates a Unix timestamp that is in the past.

**Configurations**

There are no configurations for this transformer.

**Examples**=

| Example Output |
| -------------- |
| 2524608000     |
| 946684800      |

### Generate Username\{#generate-username}

Randomly generates a username

**Configurations**

There are no configurations for this transformer.

**Examples**

| Example Output |
| -------------- |
| jdoe           |
| lsmith         |

### Generate UTC Timestamp\{#generate-utctimestamp}

Randomly generates a UTC timestamp. in the past.

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

Anonymizes and transforms an existing email address.

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

Anonymizes and transforms an existing E164 formatted phone number.

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

Anonymizes and transforms an existing first name.

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

Anonymizes and transforms an existing float value.

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

Anonymizes and transforms an existing full name.

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

Anonymizes and transforms an existing int64 phone number.

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

Anonymizes and transforms an existing int64 value.

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

Anonymizes and transforms an existing last name.

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

Anonymizes and transforms an existing phone number that is typed as a string.

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

Anonymizes and transforms an existing string value.

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

Anonymizes and transforms an existing string value by scrambling the characters while maintaining the format based on a user provided regular expression. Letters will be replaced with letters, numbers with numbers and non-number or letter ASCII characters such as "!&\*" with other characters.

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
