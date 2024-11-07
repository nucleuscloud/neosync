---
title: 11/07 - Column Strategy
hide_table_of_contents: false
slug: /column-strategy
authors:
  - evis
---

1. Added support for new column strategy to auto-map new columns
2. Added support for an IP Address Transformer
3. Added support for batching config options in the CLI
4. Added support for batching config options in the App
5. Added support for configuring max in flight for sql insert/update
6. Added support for importing and exporting mappings from the Jobs page
7. Added support for a business name transformer
8. Added a countdown component to trials to show time remaining
9. Updated sql to json datatypes in AWS output when syncing
10. Fixed a bug that was causing nil pointer errors when there were empty tables and we were syncing S3 -> database
