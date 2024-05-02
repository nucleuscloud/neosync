import { ConnectionConfig } from '@neosync/sdk';

export type ConnectionType = 'postgres' | 'mysql' | 'aws-s3' | 'openai';

export const DESTINATION_ONLY_CONNECTION_TYPES = new Set<ConnectionType>([
  'aws-s3',
]);

export function getConnectionType(
  connectionConfig: ConnectionConfig
): ConnectionType | null {
  switch (connectionConfig.config.case) {
    case 'pgConfig':
      return 'postgres';
    case 'mysqlConfig':
      return 'mysql';
    case 'awsS3Config':
      return 'aws-s3';
    case 'openaiConfig':
      return 'openai';
    default:
      return null;
  }
}
