import { TransformerConfigSchema } from '@/app/(mgmt)/[account]/new/transformer/schema';
import { JobMappingTransformer, TransformerConfig } from '@neosync/sdk';
import * as Yup from 'yup';

// Yup schema form JobMappingTransformers
export const JobMappingTransformerForm = Yup.object({
  source: Yup.number().required('A valid transformer source must be specified'),
  config: TransformerConfigSchema,
});

// Simplified version of a job mapping transformer for use with react-hook-form only
export type JobMappingTransformerForm = Yup.InferType<
  typeof JobMappingTransformerForm
>;

export function convertJobMappingTransformerToForm(
  jmt: JobMappingTransformer
): JobMappingTransformerForm {
  return {
    source: jmt.source,
    config: convertTransformerConfigToForm(jmt.config),
  };
}
export function convertJobMappingTransformerFormToJobMappingTransformer(
  form: JobMappingTransformerForm
): JobMappingTransformer {
  return new JobMappingTransformer({
    source: form.source,
    config: convertTransformerConfigSchemaToTransformerConfig(form.config),
  });
}

export function convertTransformerConfigToForm(
  tc?: TransformerConfig
): TransformerConfigSchema {
  const config = tc?.config ?? { case: '', value: {} };
  if (!config.case) {
    return { case: '', value: {} };
  }
  return {
    case: config.case,
    value: config.value.toJson(),
  };
}

export function convertTransformerConfigSchemaToTransformerConfig(
  tcs: TransformerConfigSchema
): TransformerConfig {
  // hack job that fixes bigint json transformation until we can fit this with better types
  const value = tcs.value ?? {};
  Object.entries(tcs.value).forEach(([key, val]) => {
    value[key] = val;
    if (typeof val === 'bigint') {
      value[key] = val.toString();
    }
  });
  return tcs instanceof TransformerConfig
    ? tcs
    : TransformerConfig.fromJson({
        [tcs.case ?? '']: tcs.value,
      });
}

export const JobMappingFormValues = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  column: Yup.string().required(),
  transformer: JobMappingTransformerForm,
}).required();
export type JobMappingFormValues = Yup.InferType<typeof JobMappingFormValues>;

const VIRTUAL_FOREIGN_KEY_SCHEMA = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  columns: Yup.array().of(Yup.string().required()),
}).required();

const VirtualForeignConstraintFormValues = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  columns: Yup.array().of(Yup.string().required()),
  foreignKey: VIRTUAL_FOREIGN_KEY_SCHEMA,
}).required();
export type VirtualForeignConstraintFormValues = Yup.InferType<
  typeof VirtualForeignConstraintFormValues
>;

export const SchemaFormValues = Yup.object({
  mappings: Yup.array().of(JobMappingFormValues).required(),
  virtualForeignKeys: Yup.array().of(VirtualForeignConstraintFormValues),
  connectionId: Yup.string().required(),
});
export type SchemaFormValues = Yup.InferType<typeof SchemaFormValues>;

export const SourceFormValues = Yup.object({
  sourceId: Yup.string().required('Source is required').uuid(),
  sourceOptions: Yup.object({
    haltOnNewColumnAddition: Yup.boolean().optional(),
  }),
});

export const DestinationFormValues = Yup.object({
  connectionId: Yup.string()
    .required('Connection is required')
    .uuid()
    .test(
      'checkConnectionUnique',
      'Destination must be different from source.',
      (value, ctx) => {
        if (ctx.from) {
          const { sourceId } = ctx.from[ctx.from.length - 1].value;
          if (value === sourceId) {
            return false;
          }
        }

        return true;
      }
    ),
  destinationOptions: Yup.object({
    truncateBeforeInsert: Yup.boolean().optional(),
    truncateCascade: Yup.boolean().optional(),
    initTableSchema: Yup.boolean().optional(),
    onConflictDoNothing: Yup.boolean().optional(),
  }),
}).required();
export type DestinationFormValues = Yup.InferType<typeof DestinationFormValues>;

// Intended to be used with the Source form values
export const DestinationOptionFormValues = Yup.object({
  destinationId: Yup.string().required(),

  dynamoDb: Yup.object({
    tableMappings: Yup.array()
      .of(
        Yup.object({
          sourceTable: Yup.string().required(),
          destinationTable: Yup.string().required(),
        }).required()
      )
      .required(),
  }).optional(),
});
export type DestinationOptionFormValues = Yup.InferType<
  typeof DestinationOptionFormValues
>;

export const DataSyncSourceFormValues = SourceFormValues.concat(
  SchemaFormValues
).concat(
  Yup.object({
    destinationOptions: Yup.array()
      .of(DestinationOptionFormValues.required())
      .required(),
  })
);
export type DataSyncSourceFormValues = Yup.InferType<
  typeof DataSyncSourceFormValues
>;
