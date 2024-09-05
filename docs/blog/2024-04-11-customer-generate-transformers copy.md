---
title: 4/11 - Custom Generate Transformers
hide_table_of_contents: false
slug: /custom-generate-transformers
authors:
  - evis
---

1. Add the ability to write custom javascript Transformers for a data generate job
2. Transformers now filter by data type in the schema page to prevent type mismatches
3. Add in self-service sign ups for the free tier
4. Reduced data generation and sync time by 50%
5. Logs in runs now are persistent, searchable and filterable
6. Support for setting max connection limits for connections and pooled connections
7. Added support for `ON CONFLICT` to handle duplicate rows
8. Updated retry policy to fail fast instead of waiting to fail leading to better performance
9. Update schedules to ensure that they don't get culled by Temporal
10. Support the ability to delete, cancel or terminate runs from the Jobs page instead of having to go to Runs
11. Various UI/UX refinements
