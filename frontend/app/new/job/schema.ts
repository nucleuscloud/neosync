import { IsJobNameAvailableResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import cron from 'cron-validate';
import * as Yup from 'yup';

export const DEFINE_FORM_SCHEMA = Yup.object({
  jobName: Yup.string().trim().required('Name is a required field').min(3).max(30).test(
    'checkNameUnique',
    'This name is already taken.',
    async (value) => {
      if (!value || value.length == 0 ) {
        return false;
      }
      
      const res = await isJobNameAvailable(
       value 
      );
      return res.isAvailable;
    }
  ),
  cronSchedule: Yup.string()
    .optional()
    .test('isValidCron', 'Not a valid cron schedule', (value) => {
      if (!value) {
        return true;
      }
      return !!value && cron(value).isValid();
    }),
    haltOnNewColumnAddition: Yup.boolean(),
});

export type DefineFormValues = Yup.InferType<typeof DEFINE_FORM_SCHEMA>;

export const FLOW_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid("source is required").required(),
  destinationId: Yup.string().uuid("destination is required").required(),
  // destinationIds: Yup.array().of(Yup.string().required()).required(),
});
export type FlowFormValues = Yup.InferType<typeof FLOW_FORM_SCHEMA>;

const JOB_MAPPING_SCHEMA = Yup.object({
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

export const FORM_SCHEMA = Yup.object({
  define: DEFINE_FORM_SCHEMA,
  flow: FLOW_FORM_SCHEMA,
  schema: SCHEMA_FORM_SCHEMA,
});

export type FormValues = Yup.InferType<typeof FORM_SCHEMA>;


async function isJobNameAvailable(
  name: string,
): Promise<IsJobNameAvailableResponse> {
  const res = await fetch(`/api/jobs/is-job-name-available?name=${name}`, {
    method: 'GET',
    headers: {
      'content-type': 'application/json',
    },
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return IsJobNameAvailableResponse.fromJson(await res.json());
}
