---
title: Use cases
description: Learn about the main use-cases of Neosync from anonymizing production data to subsetting data
id: usecases
hide_title: false
slug: /usecases
---

## Introduction

Neosync is a great way to anonymize sensitive data and sync it across multiple environments for better testing, debugging and developer experience. Typically, teams will point Neosync to a snapshot of their production database and anonymize their production data to make it usable in lower level environments. This is a great way to get production-like data without the risk of security and compliance challenges.

While there are many use-cases for Neosync, we're going to focus on the main ones that we see from customers.

## Safely test your code against Production data

Many developers have experienced the pain of testing their code locally and in staging and it working well and then failing in production either through an edge case or some other bug. We've all said, "It works locally though ..." at one point or other in our careers as developers. A lot of these errors tend to come up because the data that you're using locally to test against isn't representative of production data. Production data is messy and has sharp edges and that type of messiness is really difficult to manufacture in mock data.

One of the main use cases of Neosync is to anonymize production data and generate synthetic data so that it's usable locally for developers to build and test their code against. This gets them as close to testing in production as they can possibly get without any of the security and privacy risk. Not only is it a much better developer experience it has massive customer benefits. When you're able to test with realistic data you produce more resilient applications that fail less. This directly translates into happier customers and less wasted time trying to fix bugs.

## Easily reproduce Production bugs locally

Whenever we come across a bug, the first thing that we want to do is reproduce it locally so we can start to fix it. The problem is that if you don't have great data to work with that closely matches a customer's production environment, you have to hunt for the bug in order to reproduce it. This can waste a lot of time and can result in unhappy customers and frustrated developers.

The ideal debugging process would to be reproduce the customer's data state locally and then try to execute the same action that the customer took and see if there is a bug. This is where Neosync can come in and help. Neosync can help you anonymize your production and generate synthetic data so that you can use it locally and subset the data but a customer_id or any other SQL query so that you only get that customer's data. This makes the data much easier to work with.

As a developer, this is the best developer experience you can ask for as well. You're able to see almost exactly what the customer is seeing without any of the security problems and you can quickly understand what is going on. This helps you identify and fix the bug faster and make your customers happier.

## High quality data for lower-level environments

One of the biggest sources of frustration for developers is a broken staging environment. Developers rely on the fidelity of staging environments for access to data as well as the quality of the data. Whether it's hydrating local environments or running staging CI acceptance tests, it's important to have high quality staging data.

This is where Neosync can come in. Neosync can anonymize and generate synthetic data to populate staging environments with high quality data that gives developers a great developer experience.

## Solve GDPR, DPDP, FERPA, HIPAA and more

There are many data privacy regulations such as HIPAA, GDPR, DPDP and more that require that companies in countries or industries with those regulations, protect customer data. For example, a health technology company that is collecting PHI (personal health information) is required to secure that data according to HIPAA regulations. However, this can sometimes be at odds with what developers and engineering teams need. An engineering team needs data to build and test new features however if they use production data, then their development and even local systems can be in scope of HIPAA compliance. That means that they have to protect their development and local environments the same way they would protect their production environment.

This can place a big burden on security, compliance and engineering teams and isn't the right approach. One of the use-cases that we see for Neosync is to anonymize and generate synthetic data so that you can use it locally and reduce your compliance scope while still having access to high quality data. This not only helps engineering teams but also reduces the compliance and audit scope for security and compliance teams.
