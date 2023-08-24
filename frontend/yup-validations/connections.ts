import * as Yup from 'yup';

export const POSTGRES_CONNECTION = Yup.object({
  host: Yup.string().required(),
  name: Yup.string().required(),
  user: Yup.string().required(),
  pass: Yup.string().required(),
  port: Yup.number().integer().positive().required(),
  sslMode: Yup.string().optional(),
});

export type YupPostgresConnection = Yup.InferType<typeof POSTGRES_CONNECTION>;

export const NEW_POSTGRES_CONNECTION = Yup.object({
  accountId: Yup.string().uuid().required(),
  connectionName: Yup.string().required(),
  connection: POSTGRES_CONNECTION,
});

export type YupNewPostgresConnection = Yup.InferType<
  typeof NEW_POSTGRES_CONNECTION
>;

export const EXISTING_POSTGRES_CONNECTION = Yup.object({
  id: Yup.string().uuid().required(),
  connection: POSTGRES_CONNECTION,
});

export type YupExistingPostgresConnection = Yup.InferType<
  typeof EXISTING_POSTGRES_CONNECTION
>;
