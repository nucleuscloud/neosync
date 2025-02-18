import { getErrorMessage } from '@/util/util';
import { create, MessageInitShape } from '@bufbuild/protobuf';
import {
  ConnectError,
  IsConnectionNameAvailableRequestSchema,
  IsConnectionNameAvailableResponse,
} from '@neosync/sdk';
import { UseMutateAsyncFunction } from '@tanstack/react-query';
import * as Yup from 'yup';
import { getDurationValidateFn } from './number';

/* This is the standard regular expression we assign to all or most "name" fields on the backend. */
export const RESOURCE_NAME_REGEX = /^[a-z0-9-]{3,100}$/;

const connectionNameSchema = Yup.string<string, EditConnectionFormContext>()
  .required('Connection Name is a required field.')
  .min(3, 'The Connection name must be longer than 3 characters.')
  .max(100, 'The Connection name must be less than or equal to 100 characters.')
  .required()
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
        const isConnectionNameAvailable =
          context?.options?.context?.isConnectionNameAvailable;
        if (isConnectionNameAvailable) {
          const res = await isConnectionNameAvailable(
            create(IsConnectionNameAvailableRequestSchema, {
              accountId: accountId,
              connectionName: value,
            })
          );
          if (!res.isAvailable) {
            return context.createError({
              message: 'This Connection Name is already taken.',
            });
          }
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
    .min(0, 'The Port must be greater than or equal to 0.')
    .when('host', (host, schema) =>
      host
        ? schema.required('The Port is required when there is a Host.')
        : schema
    ),
  user: Yup.string().when('host', (values, schema) => {
    const [host] = values;
    return host
      ? schema.required('The User field is required when there is a Host.')
      : schema;
  }),
  knownHostPublicKey: Yup.string(),
  privateKey: Yup.string(),
  passphrase: Yup.string(),
});

export type SshTunnelFormValues = Yup.InferType<typeof SshTunnelFormValues>;

const SqlOptionsFormValues = Yup.object({
  maxConnectionLimit: Yup.number()
    .min(-1, 'The Max Open Connection Limit cannot be less than -1')
    .max(
      1000,
      'The Max Open Connection limit must be less than or equal to 1000.'
    )
    .optional(),
  maxIdleLimit: Yup.number()
    .min(-1, 'The Max Idle Connection Limit cannot be less than -1')
    .max(1000, 'The Max Idle Connection Limit cannot be greater than 1000.')
    .optional(),
  maxOpenDuration: Yup.string()
    .optional()
    .test('duration', getDurationValidateFn()),
  maxIdleDuration: Yup.string()
    .optional()
    .test('duration', getDurationValidateFn()),
});
export type SqlOptionsFormValues = Yup.InferType<typeof SqlOptionsFormValues>;

export const ClientTlsFormValues = Yup.object({
  rootCert: Yup.string(),

  clientCert: Yup.string(),
  clientKey: Yup.string(),

  serverName: Yup.string(),
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

const UrlFormValue = Yup.string().when('$activeTab', {
  is: 'url', // Only require if activeTab is 'url'
  then: (schema) => schema.required('The Connection url is required'),
  otherwise: (schema) => schema.notRequired(),
});

const UrlEnvVarFormValue = Yup.string().when('$activeTab', {
  is: 'url-env',
  then: (schema) =>
    schema
      .required('The environment variable is required')
      .matches(
        /^USER_DEFINED_/,
        'Environment variable must start with USER_DEFINED_'
      ),
  otherwise: (schema) => schema.notRequired(),
});

const MysqlDbFormValue = Yup.object().shape({
  host: Yup.string().when('$activeTab', {
    is: 'host', // Only require if activeTab is 'host'
    then: (schema) => schema.required('The host name is required.'),
    otherwise: (schema) => schema.notRequired(),
  }),
  name: Yup.string().when('$activeTab', {
    is: 'host',
    then: (schema) => schema.required('The database name is required.'),
    otherwise: (schema) => schema.notRequired(),
  }),
  user: Yup.string().when('$activeTab', {
    is: 'host',
    then: (schema) => schema.required('The database user is required.'),
    otherwise: (schema) => schema.notRequired(),
  }),
  pass: Yup.string().when('$activeTab', {
    is: 'host',
    then: (schema) => schema.required('The database password is required.'),
    otherwise: (schema) => schema.notRequired(),
  }),
  port: Yup.number()
    .integer()
    .positive()
    .when('$activeTab', {
      is: 'host',
      then: (schema) => schema.required('A database port is required.'),
      otherwise: (schema) => schema.notRequired(),
    }),
  protocol: Yup.string().when('$activeTab', {
    is: 'host',
    then: (schema) => schema.required('The database protocol is required.'),
    otherwise: (schema) => schema.notRequired(),
  }),
});

export const MysqlFormValues = Yup.object<
  MysqlCreateConnectionFormContext | MysqlEditConnectionFormContext
>().shape({
  connectionName: connectionNameSchema,
  db: MysqlDbFormValue,
  url: UrlFormValue,
  envVar: UrlEnvVarFormValue,
  tunnel: SshTunnelFormValues,
  options: SqlOptionsFormValues,
  clientTls: ClientTlsFormValues,

  activeTab: Yup.string().oneOf<ActiveConnectionTab>([
    'url',
    'host',
    'url-env',
  ]),
});

export type MysqlFormValues = Yup.InferType<typeof MysqlFormValues>;

const PostgresDbFormValue = Yup.object().shape({
  host: Yup.string().when('$activeTab', {
    is: 'host', // Only require if activeTab is 'host'
    then: (schema) => schema.required('The host name is required.'),
    otherwise: (schema) => schema.notRequired(),
  }),
  name: Yup.string().when('$activeTab', {
    is: 'host',
    then: (schema) => schema.required('The database name is required'),
    otherwise: (schema) => schema.notRequired(),
  }),
  user: Yup.string().when('$activeTab', {
    is: 'host',
    then: (schema) => schema.required('The database user is required'),
    otherwise: (schema) => schema.notRequired(),
  }),
  pass: Yup.string().when('$activeTab', {
    is: 'host',
    then: (schema) => schema.required('The database password is required'),
    otherwise: (schema) => schema.notRequired(),
  }),
  port: Yup.number()
    .integer()
    .positive()
    .when('$activeTab', {
      is: 'host',
      then: (schema) => schema.required('The database port is required'),
      otherwise: (schema) => schema.notRequired(),
    }),
  sslMode: Yup.string().optional(),
});

export const PostgresFormValues = Yup.object<
  PostgresCreateConnectionFormContext | PostgresEditConnectionFormContext
>().shape({
  connectionName: connectionNameSchema,
  db: PostgresDbFormValue,
  url: UrlFormValue,
  envVar: UrlEnvVarFormValue,
  tunnel: SshTunnelFormValues,
  options: SqlOptionsFormValues,
  clientTls: ClientTlsFormValues,

  activeTab: Yup.string().oneOf<ActiveConnectionTab>([
    'url',
    'host',
    'url-env',
  ]),
});

export type PostgresFormValues = Yup.InferType<typeof PostgresFormValues>;

const AwsCredentialsFormValues = Yup.object({
  profile: Yup.string().optional(),
  accessKeyId: Yup.string(),
  secretAccessKey: Yup.string().optional(),
  sessionToken: Yup.string().optional(),
  fromEc2Role: Yup.boolean().optional(),
  roleArn: Yup.string().optional(),
  roleExternalId: Yup.string().optional(),
});
export type AwsCredentialsFormValues = Yup.InferType<
  typeof AwsCredentialsFormValues
>;

export const AwsFormValues = Yup.object({
  connectionName: connectionNameSchema,
  s3: Yup.object({
    bucket: Yup.string().required('The Bucket name is required.'),
    pathPrefix: Yup.string().optional(),
    region: Yup.string().optional(),
    endpoint: Yup.string().optional(),
    credentials: AwsCredentialsFormValues.optional(),
  }).required('The AWS form fields are required.'),
});

export type AwsFormValues = Yup.InferType<typeof AwsFormValues>;

export const DynamoDbFormValues = Yup.object({
  connectionName: connectionNameSchema,
  db: Yup.object({
    region: Yup.string().optional(),
    endpoint: Yup.string().optional(),
    credentials: AwsCredentialsFormValues.optional(),
  }).required('The Dynamo DB form fields are required.'),
});

export type DynamoDbFormValues = Yup.InferType<typeof DynamoDbFormValues>;

export const MssqlFormValues = Yup.object<
  MssqlCreateConnectionFormContext | MssqlEditConnectionFormContext
>().shape({
  connectionName: connectionNameSchema,
  url: UrlFormValue,
  envVar: UrlEnvVarFormValue,
  options: SqlOptionsFormValues,
  tunnel: SshTunnelFormValues,
  clientTls: ClientTlsFormValues,

  activeTab: Yup.string().oneOf<MssqlActiveConnectionTab>(['url', 'url-env']),
});

export type MssqlFormValues = Yup.InferType<typeof MssqlFormValues>;

export const GcpCloudStorageFormValues = Yup.object({
  connectionName: connectionNameSchema,
  gcp: Yup.object({
    bucket: Yup.string().required('The Bucket is required.'),
    pathPrefix: Yup.string().optional(),
  }).required('The GCP form fields are required.'),
});

export type GcpCloudStorageFormValues = Yup.InferType<
  typeof GcpCloudStorageFormValues
>;

export interface CreateConnectionFormContext {
  accountId: string;

  isConnectionNameAvailable: UseMutateAsyncFunction<
    IsConnectionNameAvailableResponse,
    ConnectError,
    MessageInitShape<typeof IsConnectionNameAvailableRequestSchema>,
    unknown
  >;
}

export type ActiveConnectionTab = 'url' | 'host' | 'url-env';
export type MssqlActiveConnectionTab = 'url' | 'url-env';

export interface MysqlCreateConnectionFormContext
  extends CreateConnectionFormContext {
  activeTab: ActiveConnectionTab;
}

export type MssqlCreateConnectionFormContext = CreateConnectionFormContext;

export interface PostgresCreateConnectionFormContext
  extends CreateConnectionFormContext {
  activeTab: ActiveConnectionTab;
}

export interface EditConnectionFormContext extends CreateConnectionFormContext {
  originalConnectionName: string;
}

export interface PostgresEditConnectionFormContext
  extends EditConnectionFormContext {
  activeTab: ActiveConnectionTab;
}

export interface MysqlEditConnectionFormContext
  extends EditConnectionFormContext {
  activeTab: ActiveConnectionTab;
}

export interface MssqlEditConnectionFormContext
  extends EditConnectionFormContext {
  activeTab: MssqlActiveConnectionTab;
}

export const OpenAiFormValues = Yup.object<
  CreateConnectionFormContext | EditConnectionFormContext
>().shape({
  connectionName: connectionNameSchema,
  sdk: Yup.object({
    url: Yup.string().required('A URL must be provided.'),
    apiKey: Yup.string().required('An API Key must be provided.'),
  }).required('The Connection details are required.'),
});

export type OpenAiFormValues = Yup.InferType<typeof OpenAiFormValues>;

export const MongoDbFormValues = Yup.object({
  connectionName: connectionNameSchema,

  url: Yup.string().required('The Url is required.'),

  clientTls: ClientTlsFormValues,
}).required('The MongoDB form fields are required.');

export type MongoDbFormValues = Yup.InferType<typeof MongoDbFormValues>;
