import {
  DESTINATION_FORM_SCHEMA,
  JOB_MAPPING_SCHEMA,
  SCHEMA_FORM_SCHEMA,
  SOURCE_FORM_SCHEMA,
} from '@/yup-validations/jobs';
import { IsJobNameAvailableResponse } from '@neosync/sdk';
import * as Yup from 'yup';

const cronRegex = new RegExp(
  '^([0-5]?\\d|\\*) \\s*([01]?\\d|2[0-3]|\\*) \\s*([0-2]?\\d|3[01]|\\*|\\?) \\s*([1-9]|1[0-2]|\\*|\\?) \\s*([0-6]|\\*|\\?)$'
);

export const DEFINE_FORM_SCHEMA = Yup.object({
  jobName: Yup.string()
    .trim()
    .required('Name is a required field')
    .min(3)
    .max(30)
    .test(
      'checkNameUnique',
      'This name is already taken.',
      async (value, context) => {
        if (!value || value.length === 0) {
          return false;
        }
        const accountId = context.options.context?.accountId;
        if (!accountId) {
          return false;
        }
        const res = await isJobNameAvailable(value, accountId);
        return res.isAvailable;
      }
    ),
  cronSchedule: Yup.string()
    .optional()
    .test('validateCron', 'Invalid cron string', (value, context) => {
      const showSchedule = context.options.context?.showSchedule;
      if (!showSchedule) {
        return true;
      } else {
        return !value || cronRegex.test(value);
      }
    }),
  initiateJobRun: Yup.boolean(),
});

export type DefineFormValues = Yup.InferType<typeof DEFINE_FORM_SCHEMA>;

export const CONNECT_FORM_SCHEMA = SOURCE_FORM_SCHEMA.concat(
  Yup.object({
    destinations: Yup.array(DESTINATION_FORM_SCHEMA).required(),
  })
);

export type ConnectFormValues = Yup.InferType<typeof CONNECT_FORM_SCHEMA>;

const SINGLE_SUBSET_FORM_SCSHEMA = Yup.object({
  schema: Yup.string().trim().required(),
  table: Yup.string().trim().required(),
  whereClause: Yup.string().trim().optional(),
});

export const SINGLE_TABLE_CONNECT_FORM_SCHEMA = Yup.object({}).concat(
  DESTINATION_FORM_SCHEMA
);
export type SingleTableConnectFormValues = Yup.InferType<
  typeof SINGLE_TABLE_CONNECT_FORM_SCHEMA
>;

export const SINGLE_TABLE_SCHEMA_FORM_SCHEMA = Yup.object({
  numRows: Yup.number().required().min(1),
  schema: Yup.string().required(),
  table: Yup.string().required(),

  mappings: Yup.array().of(JOB_MAPPING_SCHEMA).required(),
});
export type SingleTableSchemaFormValues = Yup.InferType<
  typeof SINGLE_TABLE_SCHEMA_FORM_SCHEMA
>;

export const SUBSET_FORM_SCHEMA = Yup.object({
  subsets: Yup.array(SINGLE_SUBSET_FORM_SCSHEMA).required(),
});

export type SubsetFormValues = Yup.InferType<typeof SUBSET_FORM_SCHEMA>;

const FORM_SCHEMA = Yup.object({
  define: DEFINE_FORM_SCHEMA,
  connect: CONNECT_FORM_SCHEMA,
  schema: SCHEMA_FORM_SCHEMA,
  subset: SUBSET_FORM_SCHEMA.optional(),
});

export type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

async function isJobNameAvailable(
  name: string,
  accountId: string
): Promise<IsJobNameAvailableResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/is-job-name-available?name=${name}`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return IsJobNameAvailableResponse.fromJson(await res.json());
}
