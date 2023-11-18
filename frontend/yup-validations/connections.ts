import { isConnectionNameAvailable } from '@/app/new/connection/postgres/PostgresForm';
import { getAccount } from '@/components/providers/account-provider';
import * as Yup from 'yup';

export const POSTGRES_CONNECTION = Yup.object({
  host: Yup.string().required(),
  name: Yup.string().required(),
  user: Yup.string().required(),
  pass: Yup.string().required(),
  port: Yup.number().integer().positive().required(),
  sslMode: Yup.string().optional(),
});

export const NEW_POSTGRES_CONNECTION = Yup.object({
  connectionName: Yup.string().required(),
  connection: POSTGRES_CONNECTION,
});

export const EXISTING_POSTGRES_CONNECTION = Yup.object({
  id: Yup.string().uuid().required(),
  connection: POSTGRES_CONNECTION,
});

export const SSL_MODES = [
  'disable',
  'allow',
  'prefer',
  'require',
  'verify-ca',
  'verify-full',
];

const connectionNameSchema = Yup.string()
  .required()
  .test(
    'validConnectionName',
    'Connection Name must be at least 3 characters long and can only include lowercase letters, numbers, and hyphens.',
    async (value, context) => {
      if (!value || value.length < 3) {
        return false;
      }

      const regex = /^[a-z0-9-]+$/;
      if (!regex.test(value)) {
        return context.createError({
          message:
            'Connection Name can only include lowercase letters, numbers, and hyphens.',
        });
      }

      const account = getAccount();
      if (!account) {
        return false;
      }

      // this handles the case in the update flow wehre teh connection already exists
      // we return true otherwise the isConnectionName func below will fail since it already exists
      if (value == context?.options?.context?.originalConnectionName) {
        return true;
      }

      try {
        const res = await isConnectionNameAvailable(value, account.id);
        if (!res.isAvailable) {
          return context.createError({
            message: 'This Connection Name is already taken.',
          });
        }
        return true;
      } catch (error) {
        return context.createError({
          message: 'Error validating name availability.',
        });
      }
    }
  );

export const MYSQL_CONNECTION_PROTOCOLS = ['tcp', 'sock', 'pipe', 'memory'];

export const MYSQL_FORM_SCHEMA = Yup.object({
  connectionName: connectionNameSchema,
  db: Yup.object({
    host: Yup.string().required(),
    name: Yup.string().required(),
    user: Yup.string().required(),
    pass: Yup.string().required(),
    port: Yup.number().integer().positive().required(),
    protocol: Yup.string().required(),
  }).required(),
});

export type MysqlFormValues = Yup.InferType<typeof MYSQL_FORM_SCHEMA>;

export const POSTGRES_FORM_SCHEMA = Yup.object({
  connectionName: connectionNameSchema,
  db: Yup.object({
    host: Yup.string().required(),
    name: Yup.string().required(),
    user: Yup.string().required(),
    pass: Yup.string().required(),
    port: Yup.number().integer().positive().required(),
    sslMode: Yup.string().optional(),
  }).required(),
});
export type PostgresFormValues = Yup.InferType<typeof POSTGRES_FORM_SCHEMA>;

export const AWS_FORM_SCHEMA = Yup.object({
  connectionName: connectionNameSchema,
  s3: Yup.object({
    bucketArn: Yup.string().required(),
    pathPrefix: Yup.string().optional(),
    region: Yup.string().optional(),
    endpoint: Yup.string().optional(),
    credentials: Yup.object({
      profile: Yup.string().optional(),
      accessKeyId: Yup.string(),
      secretAccessKey: Yup.string().optional(),
      sessionToken: Yup.string().optional(),
      fromEc2Role: Yup.boolean().optional(),
      roleArn: Yup.string().optional(),
      roleExternalId: Yup.string().optional(),
    }).optional(),
  }).required(),
});

export type AWSFormValues = Yup.InferType<typeof AWS_FORM_SCHEMA>;
