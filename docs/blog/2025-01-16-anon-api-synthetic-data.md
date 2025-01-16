---
title: 01/16 - Synthetic Data in Anonymization API
hide_table_of_contents: false
slug: /anon-api-synthetic-data
authors:
  - evis
# cSpell:words
---

1. Add support for column removal strategy to handle schema changes in day 2 operations
2. Add support for on conflict handling in Postgres and Mysql
3. Enhance the python SDK entry point to be easier to use
4. Add support for using Neosync Transformers in the Anonymization API with replace
5. Optimized the Subset table to be way more performant
6. Add support for sourcing the connection url from an environment variable
7. Fix the Clone Transformer form
8. Add support for creating DB schema in the init schema options
9. Added ability to bulk apply subsets
10. Optimized the GenerateCity transformer to be way faster
