ALTER TABLE
  neosync_api.transformers
ADD COLUMN IF NOT EXISTS source_case text NULL;

UPDATE neosync_api.transformers
SET source_case =
  CASE source
    WHEN 0 THEN ''
    WHEN 1 THEN 'passthrough'
    WHEN 2 THEN 'default'
    WHEN 3 THEN 'transform_javascript'
    WHEN 4 THEN 'generate_email'
    WHEN 5 THEN 'transform_email'
    WHEN 6 THEN 'generate_bool'
    WHEN 7 THEN 'generate_card_number'
    WHEN 8 THEN 'generate_city'
    WHEN 9 THEN 'generate_e164_phone_number'
    WHEN 10 THEN 'generate_first_name'
    WHEN 11 THEN 'generate_float64'
    WHEN 12 THEN 'generate_full_address'
    WHEN 13 THEN 'generate_full_name'
    WHEN 14 THEN 'generate_gender'
    WHEN 15 THEN 'generate_int64_phone_number'
    WHEN 16 THEN 'generate_int64'
    WHEN 17 THEN 'generate_random_int64'
    WHEN 18 THEN 'generate_last_name'
    WHEN 19 THEN 'generate_sha256hash'
    WHEN 20 THEN 'generate_ssn'
    WHEN 21 THEN 'generate_state'
    WHEN 22 THEN 'generate_street_address'
    WHEN 23 THEN 'generate_string_phone_number'
    WHEN 24 THEN 'generate_string'
    WHEN 25 THEN 'generate_random_string'
    WHEN 26 THEN 'generate_unixtimestamp'
    WHEN 27 THEN 'generate_username'
    WHEN 28 THEN 'generate_utctimestamp'
    WHEN 29 THEN 'generate_uuid'
    WHEN 30 THEN 'generate_zipcode'
    WHEN 31 THEN 'transform_e164_phone_number'
    WHEN 32 THEN 'transform_first_name'
    WHEN 33 THEN 'transform_float64'
    WHEN 34 THEN 'transform_full_name'
    WHEN 35 THEN 'transform_int64_phone_number'
    WHEN 36 THEN 'transform_int64'
    WHEN 37 THEN 'transform_last_name'
    WHEN 38 THEN 'transform_phone_number'
    WHEN 39 THEN 'transform_string'
    WHEN 40 THEN 'null'
    WHEN 42 THEN 'generate_categorical'
    WHEN 43 THEN 'transform_character_scramble'
    WHEN 44 THEN 'custom'
    ELSE ''
  END;

  UPDATE neosync_api.transformers
  SET source_case = ''
  WHERE source_case IS NULL;

  ALTER TABLE neosync_api.transformers
  ALTER COLUMN source_case SET NOT NULL;

  ALTER TABLE
    neosync_api.transformers
  DROP COLUMN IF EXISTS source;

  ALTER TABLE
    neosync_api.transformers
  RENAME COLUMN source_case TO source;


-- This was destructive so we need to se the default to empty string so we can revert back to the column existing in a good state
ALTER TABLE neosync_api.transformers
  ADD COLUMN IF NOT EXISTS type text not null default '';

-- Revert the job mappings back to their string equivalent

-- WITH updated_mappings AS (
--     SELECT
--         id,
--         jsonb_agg(
--             jsonb_set(
--                 obj,
--                 '{jobMappingTransformerModel,source}',
--                 to_jsonb(
--                       CASE
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 0 THEN ''
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 1 THEN 'passthrough'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 2 THEN 'default'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 3 THEN 'transform_javascript'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 4 THEN 'generate_email'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 5 THEN 'transform_email'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 6 THEN 'generate_bool'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 7 THEN 'generate_card_number'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 8 THEN 'generate_city'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 9 THEN 'generate_e164_phone_number'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 10 THEN 'generate_first_name'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 11 THEN 'generate_float64'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 12 THEN 'generate_full_address'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 13 THEN 'generate_full_name'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 14 THEN 'generate_gender'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 15 THEN 'generate_int64_phone_number'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 16 THEN 'generate_int64'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 17 THEN 'generate_random_int64'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 18 THEN 'generate_last_name'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 19 THEN 'generate_sha256hash'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 20 THEN 'generate_ssn'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 21 THEN 'generate_state'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 22 THEN 'generate_street_address'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 23 THEN 'generate_string_phone_number'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 24 THEN 'generate_string'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 25 THEN 'generate_random_string'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 26 THEN 'generate_unixtimestamp'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 27 THEN 'generate_username'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 28 THEN 'generate_utctimestamp'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 29 THEN 'generate_uuid'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 30 THEN 'generate_zipcode'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 31 THEN 'transform_e164_phone_number'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 32 THEN 'transform_first_name'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 33 THEN 'transform_float64'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 34 THEN 'transform_full_name'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 35 THEN 'transform_int64_phone_number'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 36 THEN 'transform_int64'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 37 THEN 'transform_last_name'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 38 THEN 'transform_phone_number'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 39 THEN 'transform_string'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 40 THEN 'null'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 42 THEN 'generate_categorical'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 43 THEN 'transform_character_scramble'
--                         WHEN obj->'jobMappingTransformerModel'->'source' = 44 THEN 'custom'
--                       END
--                 )::jsonb
--             )
--         ) AS new_mappings
--     FROM
--         neosync_api.jobs,
--         jsonb_array_elements(mappings) AS obj
--     GROUP BY id
-- )
-- UPDATE neosync_api.jobs
-- SET mappings = updated_mappings.new_mappings
-- FROM updated_mappings
-- WHERE neosync_api.jobs.id = updated_mappings.id;
