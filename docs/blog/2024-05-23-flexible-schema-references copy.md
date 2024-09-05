---
title: 5/23 - Flexible Schema References
hide_table_of_contents: false
slug: /flexible-schema-references
authors:
  - evis
---

1. Added support for JSONB columns in subsetting section
2. Updated error handling to be more responsive in worker
3. Fixed a bug that was causing the init schema to skip on generate jobs
4. Update Terraform provider to retrieve system transformer by source
5. Exposed a logout Url override in the helm chart
6. Added support for Float transformer precision
7. Added support for unique email addresses in email transformer
8. Fixed a bug that was causing truncate to fail before an insert when there was only a single table
9. Added support for selecting a different connection that we pull the schema from which doesn't have to be the destination
10. Various UI/UX refinements
