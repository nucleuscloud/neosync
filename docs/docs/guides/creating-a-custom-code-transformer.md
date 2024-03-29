---
title: Creating a Custom Code Transformer
id: custom-code-transformers
hide_title: false
slug: /guides/custom-code-transformers
---

## Introduction

Neosync supports the ability to write your own custom logic in javascript for a transformer. We call this the `transform_javascript` transformer. Custom code transformers take in an input value at the `value` keyword and execute your custom code against that value. Custom code transformers are also only available for sync jobs since they require an input value.

## Creating a Custom Code Transformer

In order to create a custom code transformer:

1. Navigate to the **Transformers** page and click on **+ New Transformer**.
2. You'll be brought to the new transformer page where you can select a base transformer. A base transformer serves as the blueprint for the user defined transformer. Select the **Transform Javascript** transformer.

![tj](https://assets.nucleuscloud.com/neosync/docs/customcodetransformer.png)

3. Give your transformer a **Name** and a **Description**.
4. Then move onto to the **Custom Code** section. Here you can write in custom javascript code to transform your input value. Note that the value that you will be transforming is available at the `value` keyword and is of `any` type. For example, if the input value was `john` of type string then the following code:

```javascript
return value + 'test';
```

You can also do more complicated transformations such as:

```javascript
function manipulateString(str) {
  let result = reverseString(str);

  result = capitalizeString(result);

  return result;
}
// capitalize the string
function capitalizeString(s) {
  return s.toUpperCase();
}
// reverse the string
function reverseString(s) {
  return s.split('').reverse().join('');
}

const input = manipulateString(value);

return input;
```

This would return `johntest`. The code editor comes with autocomplete for most standard methods and syntax highlighting. Lastly, we do not currently support module imports in this section. **Important**: make sure you include a return statement or the custom function will not return a value.

5. Once you are satisfied with your custom code, click on the **Validate** button to ensure that your javascript compiles and is valid. If it does not compile, we will return an **invalid** error.
   ![valid](https://assets.nucleuscloud.com/neosync/docs/validcode.png)
6. Once your code has compiled, click on **Submit** to save your custom code transformer. You can now use your custom code transformer in sync jobs.

## Creating Column Dependencies

The Custom Code Transformer also allows you to condition a value on a value from another column. Let's take a look at a basic example.

Let's say that we have a table with three fields: `id, age, shouldDouble` and with types `uuid, integer, bool` respectively. And we want to create a function that will check the value in the `shouldDouble` column and if it's true, we want to double the value in the `age` column. Our code would look something like this:

```javascript
function doubleAge(age, shouldDouble) {
  if (shouldDouble) {
    return age * 2;
  } else {
    return age;
  }
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
