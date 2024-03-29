ALTER TABLE
  neosync_api.transformers
ADD COLUMN IF NOT EXISTS source_id int NULL;

-- todo: migrate data from source to source id

ALTER TABLE
  neosync_api.transformers
DROP COLUMN IF EXISTS source;

ALTER TABLE
  neosync_api.transformers
RENAME COLUMN source_id TO source;

UPDATE neosync_api.transformers
SET source = 0
WHERE source IS NULL;


ALTER TABLE neosync_api.transformers
ALTER COLUMN source SET NOT NULL;

ALTER TABLE neosync_api.transformers
  DROP COLUMN IF EXISTS type;
