import { Connection } from '@neosync/sdk';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function getRefreshIntervalFn<T>(
  fn?: (data: T) => number
): ((data: T | undefined) => number) | undefined {
  if (!fn) {
    return undefined;
  }
  return (data) => {
    if (!data) {
      return 0;
    }
    return fn(data);
  };
}

export function getSingleOrUndefined(
  item: string | string[] | undefined
): string | undefined {
  if (!item) {
    return undefined;
  }
  const newItem = Array.isArray(item) ? item[0] : item;
  return !newItem || newItem === 'undefined' ? undefined : newItem;
}

export function splitConnections(connections: Connection[]): {
  postgres: Connection[];
  mysql: Connection[];
  s3: Connection[];
  openai: Connection[];
  mongodb: Connection[];
  gcpcs: Connection[];
} {
  const postgres: Connection[] = [];
  const mysql: Connection[] = [];
  const s3: Connection[] = [];
  const openai: Connection[] = [];
  const mongodb: Connection[] = [];
  const gcpcs: Connection[] = [];

  connections.forEach((connection) => {
    if (connection.connectionConfig?.config.case === 'pgConfig') {
      postgres.push(connection);
    } else if (connection.connectionConfig?.config.case === 'mysqlConfig') {
      mysql.push(connection);
    } else if (connection.connectionConfig?.config.case === 'awsS3Config') {
      s3.push(connection);
    } else if (connection.connectionConfig?.config.case === 'openaiConfig') {
      openai.push(connection);
    } else if (connection.connectionConfig?.config.case === 'mongoConfig') {
      mongodb.push(connection);
    } else if (
      connection.connectionConfig?.config.case === 'gcpCloudstorageConfig'
    ) {
      gcpcs.push(connection);
    }
  });

  return {
    postgres,
    mysql,
    s3,
    openai,
    mongodb,
    gcpcs,
  };
}
