import { ConnectionConfig } from '@neosync/sdk';

export type ConnectionType = 'postgres' | 'mysql' | 'aws-s3' | 'openai';
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
