import * as Yup from 'yup';

export const JOB_MAPPING_SCHEMA = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  column: Yup.string().required(),
  dataType: Yup.string().required(),
  transformer: Yup.string()
    .required('Tranformer is a required field')
    .test('isValidTransformer', 'Must specify transformer', (value) => {
      return value != '';
    }),
  exclude: Yup.boolean(),
}).required();
export type JobMappingFormValues = Yup.InferType<typeof JOB_MAPPING_SCHEMA>;

export const SCHEMA_FORM_SCHEMA = Yup.object({
  mappings: Yup.array().of(JOB_MAPPING_SCHEMA).required(),
});
export type SchemaFormValues = Yup.InferType<typeof SCHEMA_FORM_SCHEMA>;

export const SOURCE_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid('source is required').required(),
  sourceOptions: Yup.object({
    haltOnNewColumnAddition: Yup.boolean().optional(),
  }),
});
export type SourceFormValues = Yup.InferType<typeof SOURCE_FORM_SCHEMA>;

export const DESTINATION_FORM_SCHEMA = Yup.object({
  destinationId: Yup.string().uuid('destination is required').required(),
  destinationOptions: Yup.object({
    truncateBeforeInsert: Yup.boolean().optional(),
    initDbSchema: Yup.boolean().optional(),
  }),
}).required();
export type DestinationFormValues = Yup.InferType<
  typeof DESTINATION_FORM_SCHEMA
>;
