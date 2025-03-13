import {
  Connection,
  GetJobRunEventsResponse,
  GetJobRunResponse,
  JobRunStatus,
} from '@neosync/sdk';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
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
  dynamodb: Connection[];
  mssql: Connection[];
} {
  const postgres: Connection[] = [];
  const mysql: Connection[] = [];
  const s3: Connection[] = [];
  const openai: Connection[] = [];
  const mongodb: Connection[] = [];
  const gcpcs: Connection[] = [];
  const dynamodb: Connection[] = [];
  const mssql: Connection[] = [];

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
    } else if (connection.connectionConfig?.config.case === 'dynamodbConfig') {
      dynamodb.push(connection);
    } else if (connection.connectionConfig?.config.case === 'mssqlConfig') {
      mssql.push(connection);
    }
  });

  return {
    postgres,
    mysql,
    s3,
    openai,
    mongodb,
    gcpcs,
    dynamodb,
    mssql,
  };
}

const TEN_SECONDS = 10 * 1000;

export function refreshJobRunWhenJobRunning(data: GetJobRunResponse): number {
  const { jobRun } = data;
  if (!jobRun || !jobRun.status) {
    return 0;
  }
  return shouldRefreshJobRun(jobRun.status) ? TEN_SECONDS : 0;
}

export function refreshWhenJobRunning(isRunning: boolean): number {
  return isRunning ? TEN_SECONDS : 0;
}

function shouldRefreshJobRun(status?: JobRunStatus): boolean {
  return (
    status === JobRunStatus.RUNNING ||
    status === JobRunStatus.PENDING ||
    status === JobRunStatus.ERROR
  );
}

export function refreshEventsWhenEventsIncomplete(
  data: GetJobRunEventsResponse
): number {
  const { isRunComplete } = data;
  return isRunComplete ? 0 : TEN_SECONDS;
}

export type JobRunsAutoRefreshInterval = 'off' | '10s' | '30s' | '1m' | '5m';

export function onJobRunsAutoRefreshInterval(
  interval: JobRunsAutoRefreshInterval
): number {
  switch (interval) {
    case 'off':
      return 0;
    case '10s':
      return 10 * 1000;
    case '30s':
      return 30 * 1000;
    case '1m':
      return 1 * 60 * 1000;
    case '5m':
      return 5 * 60 * 1000;
    default:
      return 0;
  }
}

export function onJobRunsPaused(interval: JobRunsAutoRefreshInterval): boolean {
  return interval === 'off';
}
