#!/bin/sh

# note, this command is slightly different when run on a mac.
# this is expected to be run using the docker admin-tools image
# for whatever reason (probably version difference) the CLI flags are different.
# this may change in the future as they continue to merge these things...

temporal schedule create \
      --schedule-id "6638b7bf-7905-468f-8287-5c2df4732bf0" \
      --task-queue "sync-job" \
      --workflow-type "Workflow" \
      --cron "0 0 1 1 *" \
      --pause \
      --workflow-id "6638b7bf-7905-468f-8287-5c2df4732bf0" \
      --input '{"JobId": "6638b7bf-7905-468f-8287-5c2df4732bf0"}'

temporal schedule create \
      --schedule-id "2a5d5caa-7f09-4fdf-a4a7-6a2e341aa600" \
      --task-queue "sync-job" \
      --workflow-type "Workflow" \
      --cron "0 0 1 1 *" \
      --pause \
      --workflow-id "2a5d5caa-7f09-4fdf-a4a7-6a2e341aa600" \
      --input '{"JobId": "2a5d5caa-7f09-4fdf-a4a7-6a2e341aa600"}'
