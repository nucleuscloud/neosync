-- This migration adds a new temporary column "source_id", populates it with the integer equivalent of the string value
-- It then drops the source column and renames source_id to source.
-- This migration also converts the source key in the job mappings to its integer equivalent
ALTER TABLE
  neosync_api.transformers
ADD COLUMN IF NOT EXISTS source_id int NULL;

UPDATE neosync_api.transformers
SET source_id =
  CASE source
    WHEN 'unspecified' THEN 0
    WHEN '' THEN 0
    WHEN 'passthrough' THEN 1
    WHEN 'default' THEN 2
    WHEN 'generate_default' THEN 2
    WHEN 'transform_javascript' THEN 3
    WHEN 'generate_email' THEN 4
    WHEN 'transform_email' THEN 5
    WHEN 'generate_bool' THEN 6
    WHEN 'generate_card_number' THEN 7
    WHEN 'generate_city' THEN 8
    WHEN 'generate_e164_phone_number' THEN 9
    WHEN 'generate_first_name' THEN 10
    WHEN 'generate_float64' THEN 11
    WHEN 'generate_full_address' THEN 12
    WHEN 'generate_full_name' THEN 13
    WHEN 'generate_gender' THEN 14
    WHEN 'generate_int64_phone_number' THEN 15
    WHEN 'generate_int64' THEN 16
    WHEN 'generate_random_int64' THEN 17
    WHEN 'generate_last_name' THEN 18
    WHEN 'generate_sha256hash' THEN 19
    WHEN 'generate_ssn' THEN 20
    WHEN 'generate_state' THEN 21
    WHEN 'generate_street_address' THEN 22
    WHEN 'generate_string_phone_number' THEN 23
    WHEN 'generate_string' THEN 24
    WHEN 'generate_random_string' THEN 25
    WHEN 'generate_unixtimestamp' THEN 26
    WHEN 'generate_username' THEN 27
    WHEN 'generate_utctimestamp' THEN 28
    WHEN 'generate_uuid' THEN 29
    WHEN 'generate_zipcode' THEN 30
    WHEN 'transform_e164_phone_number' THEN 31
    WHEN 'transform_first_name' THEN 32
    WHEN 'transform_float64' THEN 33
    WHEN 'transform_full_name' THEN 34
    WHEN 'transform_int64_phone_number' THEN 35
    WHEN 'transform_int64' THEN 36
    WHEN 'transform_last_name' THEN 37
    WHEN 'transform_phone_number' THEN 38
    WHEN 'transform_string' THEN 39
    WHEN 'null' THEN 40
    WHEN 'generate_categorical' THEN 42
    WHEN 'transform_character_scramble' THEN 43
    WHEN 'user_defined' THEN 44
    WHEN 'custom' THEN 44
  END;

-- incase we missed any, just set it to unspecified to ensure there are no nulls
UPDATE neosync_api.transformers
SET source = 0
WHERE source IS NULL;

ALTER TABLE neosync_api.transformers
ALTER COLUMN source_id SET NOT NULL;

ALTER TABLE
  neosync_api.transformers
DROP COLUMN IF EXISTS source;

ALTER TABLE
  neosync_api.transformers
RENAME COLUMN source_id TO source;

ALTER TABLE neosync_api.transformers
  DROP COLUMN IF EXISTS type;

-- Update job mappings
WITH updated_mappings AS (
    SELECT
        id,
        jsonb_agg(
            jsonb_set(
                obj,
                '{jobMappingTransformerModel,source}',
                to_jsonb(
                      CASE
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'unspecified' THEN 0
                        WHEN obj->'jobMappingTransformerModel'->>'source' = '' THEN 0
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'passthrough' THEN 1
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'default' THEN 2
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_default' THEN 2
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_javascript' THEN 3
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_email' THEN 4
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_email' THEN 5
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_bool' THEN 6
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_card_number' THEN 7
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_city' THEN 8
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_e164_phone_number' THEN 9
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_first_name' THEN 10
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_float64' THEN 11
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_full_address' THEN 12
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_full_name' THEN 13
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_gender' THEN 14
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_int64_phone_number' THEN 15
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_int64' THEN 16
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_random_int64' THEN 17
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_last_name' THEN 18
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_sha256hash' THEN 19
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_ssn' THEN 20
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_state' THEN 21
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_street_address' THEN 22
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_street_address' THEN 22
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_string_phone_number' THEN 23
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_string' THEN 24
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_random_string' THEN 25
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_unixtimestamp' THEN 26
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_username' THEN 27
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_utctimestamp' THEN 28
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_uuid' THEN 29
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_zipcode' THEN 30
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_e164_phone_number' THEN 31
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_first_name' THEN 32
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_float64' THEN 33
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_full_name' THEN 34
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_int64_phone_number' THEN 35
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_int64' THEN 36
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_last_name' THEN 37
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_phone_number' THEN 38
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_string' THEN 39
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'null' THEN 40
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'generate_categorical' THEN 42
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'transform_character_scramble' THEN 43
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'user_defined' THEN 44
                        WHEN obj->'jobMappingTransformerModel'->>'source' = 'custom' THEN 44
                      END
                )
            )
        ) AS new_mappings
    FROM
        neosync_api.jobs,
        jsonb_array_elements(mappings) AS obj
    GROUP BY id
)
UPDATE neosync_api.jobs
SET mappings = updated_mappings.new_mappings
FROM updated_mappings
WHERE neosync_api.jobs.id = updated_mappings.id;
