import { isConnectionNameAvailable } from '@/app/(mgmt)/[account]/connections/util';
import { getErrorMessage } from '@/util/util';
import * as Yup from 'yup';

/* This is the standard regular expression we assign to all or most "name" fields on the backend. */
export const RESOURCE_NAME_REGEX = /^[a-z0-9-]{3,30}$/;

const connectionNameSchema = Yup.string()
  .required()
  .min(3)
  .max(30)
  .test(
    'validConnectionName',
    'Connection Name must be at least 3 characters long and can only include lowercase letters, numbers, and hyphens.',
    async (value, context) => {
      if (!value || value.length < 3) {
        return false;
      }

      if (!RESOURCE_NAME_REGEX.test(value)) {
        return context.createError({
          message:
            'Connection Name can only include lowercase letters, numbers, and hyphens.',
        });
      }

      // this handles the case in the update flow where the connection already exists
      // we return true otherwise the isConnectionName func below will fail since it already exists
      if (value === context?.options?.context?.originalConnectionName) {
        return true;
      }

      const accountId = context?.options?.context?.accountId;
      if (!accountId) {
        return context.createError({
          message: 'Account is not valid.',
        });
      }

      try {
        const res = await isConnectionNameAvailable(value, accountId);
        if (!res.isAvailable) {
          return context.createError({
            message: 'This Connection Name is already taken.',
          });
        }
        return true;
      } catch (error) {
        return context.createError({
          message: `Error: ${getErrorMessage(error)}`,
        });
      }
    }
  );

// todo: need to do better validation here
export const SshTunnelFormValues = Yup.object({
  host: Yup.string(),
  port: Yup.number()
    .min(0)
    .when('host', (host, schema) => ([host] ? schema.required() : schema)),
  user: Yup.string().when('host', (values, schema) => {
    const [host] = values;
    return host ? schema.required() : schema;
  }),

  knownHostPublicKey: Yup.string(),

  privateKey: Yup.string(),
  passphrase: Yup.string(),
});
export type SshTunnelFormValues = Yup.InferType<typeof SshTunnelFormValues>;

const SQL_OPTIONS_FORM_SCHEMA = Yup.object({
  maxConnectionLimit: Yup.number().min(0).max(10000).optional(),
});

export const ClientTlsFormValues = Yup.object({
  rootCert: Yup.string(),

  clientCert: Yup.string(),
  clientKey: Yup.string(),
});
export type ClientTlsFormValues = Yup.InferType<typeof ClientTlsFormValues>;

export const SSL_MODES = [
  'disable',
  'allow',
  'prefer',
  'require',
  'verify-ca',
  'verify-full',
];

export const MYSQL_CONNECTION_PROTOCOLS = ['tcp', 'sock', 'pipe', 'memory'];

export const MysqlFormValues = Yup.object({
  connectionName: connectionNameSchema,
  db: Yup.object({
    host: Yup.string().required(),
    name: Yup.string().required(),
    user: Yup.string().required(),
    pass: Yup.string().required(),
    port: Yup.number().integer().positive().required(),
    protocol: Yup.string().required(),
  }).required(),
  url: Yup.string().when('$activeTab', {
    is: 'url', // Only require if activeTab is 'url'
    then: (schema) => schema.required('The connection url is required'),
  }),
  tunnel: SshTunnelFormValues,
  options: SQL_OPTIONS_FORM_SCHEMA,
});

export type MysqlFormValues = Yup.InferType<typeof MysqlFormValues>;

export const POSTGRES_FORM_SCHEMA = Yup.object({
  connectionName: connectionNameSchema,
  db: Yup.object({
    host: Yup.string().when('$activeTab', {
      is: 'parameters', // Only require if activeTab is 'parameters'
      then: (schema) => schema.required('The host name is required.'),
      otherwise: (schema) => schema.notRequired(),
    }),
    name: Yup.string().when('$activeTab', {
      is: 'parameters',
      then: (schema) => schema.required('The database name is required'),
      otherwise: (schema) => schema.notRequired(),
    }),
    user: Yup.string().when('$activeTab', {
      is: 'parameters',
      then: (schema) => schema.required('The database user is required'),
      otherwise: (schema) => schema.notRequired(),
    }),
    pass: Yup.string().when('$activeTab', {
      is: 'parameters',
      then: (schema) => schema.required('The database password is required'),
      otherwise: (schema) => schema.notRequired(),
    }),
    port: Yup.number()
      .integer()
      .positive()
      .when('$activeTab', {
        is: 'parameters',
        then: (schema) => schema.required('The database port is required'),
        otherwise: (schema) => schema.notRequired(),
      }),
    sslMode: Yup.string().optional(),
  }),
  url: Yup.string().when('$activeTab', {
    is: 'url', // Only require if activeTab is 'url'
    then: (schema) => schema.required('The connection url is required'),
  }),
  tunnel: SshTunnelFormValues,
  options: SQL_OPTIONS_FORM_SCHEMA,
  clientTls: ClientTlsFormValues,
});

export type PostgresFormValues = Yup.InferType<typeof POSTGRES_FORM_SCHEMA>;

export const AWS_FORM_SCHEMA = Yup.object({
  connectionName: connectionNameSchema,
  s3: Yup.object({
    bucket: Yup.string().required(),
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

export const GcpCloudStorageFormValues = Yup.object({
  connectionName: connectionNameSchema,
  gcp: Yup.object({
    bucket: Yup.string().required(),
    pathPrefix: Yup.string().optional(),
  }).required(),
});

export type GcpCloudStorageFormValues = Yup.InferType<
  typeof GcpCloudStorageFormValues
>;

export interface CreateConnectionFormContext {
  accountId: string;
}
export interface EditConnectionFormContext extends CreateConnectionFormContext {
  originalConnectionName: string;
}

export const OpenAiFormValues = Yup.object({
  connectionName: connectionNameSchema,
  sdk: Yup.object({
    url: Yup.string().required('A URL must be provided.'),
    apiKey: Yup.string().required('An API Key must be provided.'),
  }).required(),
});

export type OpenAiFormValues = Yup.InferType<typeof OpenAiFormValues>;

export const MongoDbFormValues = Yup.object({
  connectionName: connectionNameSchema,

  url: Yup.string().required(),

  clientTls: ClientTlsFormValues,
}).required();

export type MongoDbFormValues = Yup.InferType<typeof MongoDbFormValues>;
