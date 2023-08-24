import cron from 'cron-validate';
import * as Yup from 'yup';

export const DEFINE_FORM_SCHEMA = Yup.object({
  jobName: Yup.string().required('Name is a required field').min(3).max(30),
  cronSchedule: Yup.string()
    // .required('Cron Schedule is a required field')
    .optional()
    .test('isValidCron', 'Not a valid cron schedule', (value) => {
      return !!value && cron(value).isValid();
    }),
});

export type DefineFormValues = Yup.InferType<typeof DEFINE_FORM_SCHEMA>;

export const FLOW_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid().required(),
  destinationId: Yup.string().uuid().required(),
  // destinationIds: Yup.array().of(Yup.string().required()).required(),
});
export type FlowFormValues = Yup.InferType<typeof FLOW_FORM_SCHEMA>;

export const SCHEMA_FORM_SCHEMA = Yup.object();
export type SchemaFormValues = Yup.InferType<typeof SCHEMA_FORM_SCHEMA>;

export const FORM_SCHEMA = Yup.object({
  define: DEFINE_FORM_SCHEMA,
  flow: FLOW_FORM_SCHEMA,
  schema: SCHEMA_FORM_SCHEMA,
});

export type FormValues = Yup.InferType<typeof FORM_SCHEMA>;
