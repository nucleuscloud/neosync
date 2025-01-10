---
title: SQL Javascript Transformer Types
slug: /transformers/sql-javascript
hide_title: false
id: sql-javascript
description: Learn about SQL types in the javascript transformer
---

# SQL to JavaScript Type Mapping

This guide provides a comprehensive overview of SQL to JavaScript type mappings for PostgreSQL, MySQL, and SQL Server databases. It includes detailed tables showing how various SQL data types are represented in JavaScript, along with examples of common date and string manipulation techniques in JavaScript for working with database data.

<!-- cspell:disable  -->

## PostgreSQL

| SQL Type         | JavaScript Type | Example Value                                         |
| ---------------- | --------------- | ----------------------------------------------------- |
| smallint         | number          | 32767                                                 |
| integer          | number          | 2147483647                                            |
| bigint           | number          | 9223372036854776000                                   |
| decimal          | string          | "1234.56"                                             |
| numeric          | string          | "99999999.99"                                         |
| real             | number          | 12345.669921875                                       |
| double precision | number          | 3.14159265                                            |
| serial           | number          | 1                                                     |
| bigserial        | number          | 1                                                     |
| money            | string          | "$100.00"                                             |
| char             | string          | "A "                                                  |
| varchar          | string          | "Example varchar"                                     |
| text             | string          | "Example text"                                        |
| bytea            | neosync.Binary  |[Binary type](/transformers/neosync-types#binary)    |
| timestamp        | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                                   |
| timestamptz      | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                                   |
| date             | neosync.NeosyncDateTime          | [NeosyncDateTime ](/transformers/neosync-types#neosyncdatetime)                                   |
| time             | string          | "12:34:56"                                            |
| timetz           | string          | "12:34:56+00"                                         |
| interval         | neosync.Interval| [Interval type](/transformers/neosync-types#interval)                                      |
| boolean          | boolean         | true                                                  |
| point            | string          | "(1,2)"                                               |
| line             | string          | "{1,1,0}"                                             |
| lseg             | string          | "[(0,0),(1,1)]"                                       |
| box              | string          | "(1,1),(0,0)"                                         |
| path             | string          | "((0,0),(1,1),(2,2))"                                 |
| polygon          | string          | "((0,0),(1,1),(1,0))"                                 |
| circle           | string          | `<(1,1),1>`                                           |
| cidr             | string          | "192.168.1.0/24"                                      |
| inet             | string          | "192.168.1.1"                                         |
| macaddr          | string          | "08:00:2b:01:02:03"                                   |
| bit              | neosync.Bits    | [Bits type](/transformers/neosync-types#bits)         |
| varbit           | neosync.Bits    | [Bits type](/transformers/neosync-types#bits)         |
| tsvector         | string          | "'example' 'tsvector'"                                |
| uuid             | string          | "123e4567-e89b-12d3-a456-426614174000"                |
| xml              | string          | `<foo>bar</foo>`                                      |
| json             | object          | `{"key": "value", "array": [1, 2, 3]}`                |
| jsonb            | object          | `{"key": "value", "array": [1, 2, 3]}`                |
| int4range        | string          | "[1,11)"                                              |
| int8range        | string          | "[1,1001)"                                            |
| numrange         | string          | "[1.0,10.0]"                                          |
| tsrange          | string          | '["2024-01-01 12:00:00","2024-01-01 13:00:00"]'       |
| tstzrange        | string          | '["2024-01-01 12:00:00+00","2024-01-01 13:00:00+00"]' |
| daterange        | string          | "[2024-01-01,2024-01-03)"                             |
| oid              | number          | 123456                                                |
| text[]           | array           | ["cat","dog"]                                         |

## MySQL

| SQL Type   | JavaScript Type | Example Value                          |
| ---------- | --------------- | -------------------------------------- |
| tinyint    | number          | 127                                    |
| smallint   | number          | 32767                                  |
| mediumint  | number          | 8388607                                |
| int        | number          | 2147483647                             |
| bigint     | number          | 92233720368547710                      |
| decimal    | string          | "1234.56"                              |
| float      | number          | 3.1414999961853027                     |
| double     | number          | 3.14159265                             |
| bit        | neosync.Bits    | [Bits type](/transformers/neosync-types#bits)         |
| char       | string          | "Fixed Char"                           |
| varchar    | string          | "Variable Char"                        |
| binary     | neosync.Binary  | [Binary type](/transformers/neosync-types#binary)    |
| varbinary  | neosync.Binary  | [Binary type](/transformers/neosync-types#binary)    |
| tinyblob   | string          | "Tiny BLOB"                            |
| blob       | string          | "Regular BLOB"                         |
| mediumblob | string          | "Medium BLOB"                          |
| longblob   | string          | "Long BLOB"                            |
| tinytext   | string          | "Tiny Text"                            |
| text       | string          | "Regular Text"                         |
| mediumtext | string          | "Medium Text"                          |
| longtext   | string          | "Long Text"                            |
| enum       | string          | "value2"                               |
| set        | string          | "option1,option3"                      |
| json       | object          | `{"key": "value", "array": [1, 2, 3]}` |
| date       | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| time       | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| datetime   | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| timestamp  | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| year       | number          | 2023                                   |

## SQL Server

| SQL Type         | JavaScript Type | Example Value |
| ---------------- | --------------- | --------------- |
| bit              | boolean         | true                                       |
| tinyint          | number          | 255                                        |
| smallint         | number          | 32767                                      |
| int              | number          | 2147483647                                 |
| bigint           | number          | 9223372036854776000                        |
| decimal          | string          | "1234.56"                                  |
| numeric          | string          | "1234.56"                                  |
| smallmoney       | string          | "1234.5600"                                |
| money            | string          | "1234.5678"                                |
| float            | number          | 3.14159                                    |
| real             | number          | 3.140000104904175                          |
| date             | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| time             | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| datetime         | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| datetime2        | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| smalldatetime    | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| datetimeoffset   | neosync.NeosyncDateTime          | [NeosyncDateTime](/transformers/neosync-types#neosyncdatetime)                    |
| binary         | neosync.Binary| [Binary type](/transformers/neosync-types#binary)                                      |
| varbinary         | neosync.Bits| [Bits type](/transformers/neosync-types#bits)                                      |
| char             | string          | "CHAR"                                     |
| varchar          | string          | "VARCHAR"                                  |
| varchar(max)     | string          | "VARCHAR(MAX)"                             |
| text             | string          | "TEXT"                                     |
| nchar            | string          | "NCHAR "                                   |
| nvarchar         | string          | "NVARCHAR"                                 |
| nvarchar(max)    | string          | "NVARCHAR(MAX)"                            |
| ntext            | string          | "NTEXT"                                    |
| xml              | string          | `<root><element>XML Data</element></root>` |
| uniqueidentifier | string          | "2405c0f1-61fa-ce4f-b49f-df6414d3b502"     |
| sql_variant      | string          | "SQL_VARIANT"                              |

<!-- cspell:enable -->

## Manipulating Dates in JavaScript

When working with dates from SQL databases in JavaScript, you often need to perform various manipulations. Here are some common date manipulation techniques:

> SQL timestamp and date columns are a special objects in Javascript that need to be converted to Javascript Date types.
> <br/>

```javascript
new Date(value.UnixNano() / 1e6);
```

> Timestamps and Dates can be returned as either a Javascript Date or a correctly formatted date string.
> <br/>

```javascript
return new Date('2023-09-20T14:30:00Z');
// OR
return '2023-09-20T14:30:00Z';
```

**Creating Date Objects**:

```javascript
// Convert SQL date object to JS Date
const inputDate = new Date(value.UnixNano() / 1e6);

// From a string
const date1 = new Date('2023-09-20T14:30:00Z');

// From year, month (0-11), day, hour, minute, second
const date2 = new Date(2023, 8, 20, 14, 30, 0); // September 20, 2023, 14:30:00

// Current date and time
const now = new Date();
```

**Getting Date Components**:

```javascript
const date = new Date('2023-09-20T14:30:00Z');
console.log(date.getFullYear()); // 2023
console.log(date.getMonth()); // 8 (0-11, so 8 is September)
console.log(date.getDate()); // 20
console.log(date.getHours()); // 14
console.log(date.getMinutes()); // 30
console.log(date.getSeconds()); // 0
```

**Setting Date Components**:

```javascript
const date = new Date('2023-09-20T14:30:00Z');
date.setFullYear(2024);
date.setMonth(0); // January
date.setDate(15);
date.setHours(10);
date.setMinutes(45);
date.setSeconds(30);
console.log(date); // 2024-01-15T10:45:30.000Z
```

**Adding or Subtracting Time**:

```javascript
const date = new Date('2023-09-20T14:30:00Z');

// Add one day
date.setDate(date.getDate() + 1);
console.log(date); // 2023-09-21T14:30:00.000Z

// Subtract 2 hours
date.setHours(date.getHours() - 2);
console.log(date); // 2023-09-21T12:30:00.000Z

// Add 30 minutes
date.setMinutes(date.getMinutes() + 30);
console.log(date); // 2023-09-21T13:00:00.000Z
```

Remember that JavaScript uses zero-based indexing for months (0-11), so January is 0 and December is 11. Also, be aware of time zone differences when working with dates, especially when dealing with data from SQL databases in different time zones.

## Manipulating Strings in JavaScript

> String interpolation or template literals are unsupported in the transformer. Instead use string concatenation.
> <br/>

```javascript
// Unsupported
const str = `Hello, ${name}!`;

// Supported
const str = 'Hello, ' + name + '!';
```

**Concatenation**:

```javascript
var firstName = 'John';
var lastName = 'Doe';
var fullName = firstName + ' ' + lastName;
console.log(fullName); // "John Doe"
```

**Substring Extraction**:

```javascript
var text = 'Hello, World!';
var hello = text.substring(0, 5);
var world = text.slice(7, 12);
console.log(hello); // "Hello"
console.log(world); // "World"
```

**Changing Case**:

```javascript
var text = 'Hello, World!';
var upperCase = text.toUpperCase();
var lowerCase = text.toLowerCase();
console.log(upperCase); // "HELLO, WORLD!"
console.log(lowerCase); // "hello, world!"
```

**Trimming Whitespace**:

```javascript
var text = '   Hello, World!   ';
var trimmed = text.trim();
console.log(trimmed); // "Hello, World!"
```

**Replacing Substrings**:

```javascript
var text = 'Hello, World!';
var newText = text.replace('World', 'JavaScript');
console.log(newText); // "Hello, JavaScript!"
```

**Splitting Strings**:

```javascript
var csvData = 'John,Doe,30,New York';
var dataArray = csvData.split(',');
console.log(dataArray); // ["John", "Doe", "30", "New York"]
```

**Checking for Substrings**:

```javascript
var text = 'Hello, World!';
var containsHello = text.includes('Hello');
var startsWithHello = text.startsWith('Hello');
var endsWithWorld = text.endsWith('World!');
console.log(containsHello); // true
console.log(startsWithHello); // true
console.log(endsWithWorld); // true
```

**Finding Substring Index**:

```javascript
var text = 'Hello, World!';
var indexOfWorld = text.indexOf('World');
console.log(indexOfWorld); // 7
```

**Repeating Strings**:

```javascript
var star = '*';
var stars = star.repeat(5);
console.log(stars); // "*****"
```

**Padding Strings**:
`javascript
    var number = "42";
    var paddedNumber = number.padStart(5, "0");
    console.log(paddedNumber);  // "00042"
    `

**Extracting Characters**:
`javascript
    var text = "Hello";
    var firstChar = text.charAt(0);
    var lastChar = text.charAt(text.length - 1);
    console.log(firstChar);  // "H"
    console.log(lastChar);   // "o"
    `

## Manipulating Objects

**Creating Objects**

```javascript
// Object literal notation
let person = {
  name: 'John Doe',
  age: 30,
  job: 'Developer',
};

// Using the Object constructor
let car = new Object();
car.make = 'Toyota';
car.model = 'Corolla';
car.year = 2022;
```

**Accessing Object Properties**

```javascript
console.log(person.name); // "John Doe"
console.log(car['model']); // "Corolla"
```

**Object.keys(), Object.values(), and Object.entries()**

```javascript
let keys = Object.keys(person);
let values = Object.values(person);
let entries = Object.entries(person);
```

## Manipulating Arrays

**Creating Arrays**

```javascript
let fruits = ['apple', 'banana', 'orange'];
let numbers = new Array(1, 2, 3, 4, 5);
```

**Accessing Array Elements**

```javascript
console.log(fruits[0]); // "apple"
console.log(fruits[fruits.length - 1]); // "orange" (last element)
```

**Adding and Removing Elements**

```javascript
fruits.push('grape'); // Add to end
fruits.unshift('mango'); // Add to beginning
let lastFruit = fruits.pop(); // Remove from end
let firstFruit = fruits.shift(); // Remove from beginning
```

**Iterating**

```javascript
fruits.forEach((fruit) => console.log(fruit));

let upperFruits = fruits.map((fruit) => fruit.toUpperCase());

let longFruits = fruits.filter((fruit) => fruit.length > 5);
```

**Finding Elements**

```javascript
let hasApple = fruits.includes('apple');
let bananaIndex = fruits.indexOf('banana');
let fruit = fruits.find((f) => f.startsWith('o'));
```

\*\*
