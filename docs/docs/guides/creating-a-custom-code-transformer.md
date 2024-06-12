---
title: Creating a Custom Code Transformer
description: Learn how to create a custom code transformer that you can use to implement your own anonymization and synthetic data logic
id: custom-code-transformers
hide_title: false
slug: /guides/custom-code-transformers
# cSpell:words goja Goja
---

## Introduction

Neosync supports the ability to write your own custom logic using JavaScript.

There exist two different transformers that enable this. One of them is input-free, while the other is input-full, and allows you to transform incoming values.

- `transform_javascript` takes in and allows you to modify input. This transformer may be used in Sync jobs.
- `generate_javascript` takes in no input and only expects an output. This transformer may be used in both Sync and Generate jobs.

## Creating a Custom Code Transformer

There exists two ways to configure a javascript transformer:

1. Creating a custom transformer and saving it to then later be used when configuring the job's column mappings.
2. Selecting Generate/Transformer Javascript and putting the code directly in line for that specific job's column mapping(s).

This guide will walk you through the first option.

In order to create a custom code transformer:

1. Navigate to the **Transformers** page and click on **+ New Transformer**.
2. You'll be brought to the new transformer page where you can select a base transformer. A base transformer serves as the blueprint for the user defined transformer. Select the **Transform Javascript** transformer.

![tj](https://assets.nucleuscloud.com/neosync/docs/customcodetransformer.png)

3. Give your transformer a **Name** and a **Description**.
4. Then move onto to the **Custom Code** section. Here you can write in custom javascript code to transform your input value.
   Note that the value that you will be transforming is available at the `value` keyword and is coerced into the relevant javascript type.

For example, if the input value was `john` with a database type `TEXT` then the following code: `return value + 'test';` would return `johntest`.

The code editor comes with autocomplete for most standard methods and syntax highlighting.

**The JavaScript transformers do not currently support 3rd party modules. If necessary, you must bundle them directly into the code.**

**Important**: make sure you include a return statement or the custom function will not return a value.

An example of a more complicated transformation:

```javascript
function manipulateString(s) {
  return capitalizeString(reverseString(s));
}
// capitalize the string
function capitalizeString(s) {
  return s.toUpperCase();
}
// reverse the string
function reverseString(s) {
  return s.split('').reverse().join('');
}
return manipulateString(value);
```

5. Once you are satisfied with your custom code, click on the **Validate** button to ensure that your javascript compiles and is valid. If it does not compile, we will return an **invalid** error. Please be aware that just because the program compiles _does not mean_ that it will run. It simply validates that you are not using any invalid JS keywords. You may still very well run into a runtime error!
   ![valid](/img/valid-javascript.png)
6. Once your code has compiled, click on **Submit** to save your custom code transformer. You can now use your custom code transformer in sync jobs. The custom transformer will appear under the user defined/custom section of the Transformer drop down in the job mapping table.

## Creating Column Dependencies

The Custom Code Transformer also allows you to condition a value on a value from another column. Let's take a look at a basic example.

Let's say that we have a table with three fields: `id, age, shouldDouble` and with types `uuid, integer, bool` respectively. And we want to create a function that will check the value in the `shouldDouble` column and if it's true, we want to double the value in the `age` column. Our code would look something like this:

```javascript
function doubleAge(age, shouldDouble) {
  return shouldDouble ? age * 2 : age;
}
return doubleAge(value, input.shouldDouble);
```

The key thing to point out here is the `input.shouldDouble` argument we're passing into the `return doubleAge(...)` function. We can use the `input.{column_name}` notation to access values from other columns.

For example, if we had a table that looked like this:

| age | shouldDouble |
| --- | ------------ |
| 20  | true         |

Our output to the `age` column would be: `40`

## Things to watch out for

### Interpolations

There are a few thing to watch out for as you're writing your custom code. One is that interpolations are a work in progress. For example:

```javascript
const iam = 'I am';
const five = 5;
return `${iam} ${five}`;
```

Won't work appropriately since the underlying system we use to compile and run the javascript doesn't recognize that. Instead you would have to do something like:

```javascript
const iam = 'I am';
const five = 5;
return hello + five.toString();
```

Calling the `toString()` method on the integer in order to return it correctly. We're working on this and will update it once we have a fix.

## JavaScript Runtime

Since Neosync is written in Go, we have to use a JavaScript-compatible runtime to invoke JS transformations. Today, we rely on [goja](https://github.com/dop251/goja). This works well and is written in native Golang. However, it does have limitations such as no 3rd party modules, and does not have full ES6 support (yet).

If running into anything strange with JS, it's worthwhile to check what Goja supports to see if you've hit any incompatibility errors.
