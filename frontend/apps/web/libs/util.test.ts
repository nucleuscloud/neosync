/**
 * @jest-environment node
 */

import {
  AwsS3ConnectionConfig,
  Connection,
  ConnectionConfig,
  DynamoDBConnectionConfig,
  GcpCloudStorageConnectionConfig,
  GetJobRunEventsResponse,
  GetJobRunResponse,
  JobRun,
  JobRunStatus,
  MongoConnectionConfig,
  MysqlConnectionConfig,
  OpenAiConnectionConfig,
  PostgresConnectionConfig,
} from '@neosync/sdk';
import {
  getSingleOrUndefined,
  JobRunsAutoRefreshInterval,
  onJobRunsAutoRefreshInterval,
  onJobRunsPaused,
  refreshEventsWhenEventsIncomplete,
  refreshJobRunWhenJobRunning,
  splitConnections,
} from './utils';

describe('getSingleOrUndefined', () => {
  it('should return a string for a string', () => {
    const input = 'test';
    const result = getSingleOrUndefined(input);
    expect(result).toBe('test');
  });
  it('should return a string for a string array', () => {
    const input = ['hello'];
    const result = getSingleOrUndefined(input);
    expect(result).toBe('hello');
  });
  it('should return undefined for an empty array', () => {
    const input = [''];
    const result = getSingleOrUndefined(input);
    expect(result).toBeUndefined();
  });
  it('should return undefined or undefined for undefined', () => {
    const input = undefined;
    const result = getSingleOrUndefined(input);
    expect(result).toBe(undefined);
  });
  it('should return undefined for an array with undefined as an element', () => {
    const input = ['undefined'];
    const result = getSingleOrUndefined(input);
    expect(result).toBeUndefined();
  });
  it('should return undefined for a string undefined', () => {
    const input = 'undefined';
    const result = getSingleOrUndefined(input);
    expect(result).toBeUndefined();
  });
});

describe('splitConnections', () => {
  const postgres = new Connection({
    connectionConfig: {
      config: { case: 'pgConfig', value: {} as PostgresConnectionConfig },
    } as ConnectionConfig,
  });
  const mysql = new Connection({
    connectionConfig: {
      config: { case: 'mysqlConfig', value: {} as MysqlConnectionConfig },
    } as ConnectionConfig,
  });
  const mssql = new Connection({
    connectionConfig: {
      config: { case: 'mssqlConfig', value: {} as MssqlConnectionConfig },
    } as ConnectionConfig,
  });
  const s3 = new Connection({
    connectionConfig: {
      config: { case: 'awsS3Config', value: {} as AwsS3ConnectionConfig },
    } as ConnectionConfig,
  });
  const openai = new Connection({
    connectionConfig: {
      config: { case: 'openaiConfig', value: {} as OpenAiConnectionConfig },
    } as ConnectionConfig,
  });
  const mongodb = new Connection({
    connectionConfig: {
      config: { case: 'mongoConfig', value: {} as MongoConnectionConfig },
    } as ConnectionConfig,
  });
  const gcpcs = new Connection({
    connectionConfig: {
      config: {
        case: 'gcpCloudstorageConfig',
        value: {} as GcpCloudStorageConnectionConfig,
      },
    } as ConnectionConfig,
  });
  const dynamodb = new Connection({
    connectionConfig: {
      config: {
        case: 'dynamodbConfig',
        value: {} as DynamoDBConnectionConfig,
      },
    } as ConnectionConfig,
  });
  it('should return an object with only a postgres connection as a property', () => {
    const connections: Connection[] = [postgres];
    const result = splitConnections(connections);
    const expected = {
      postgres: [postgres],
      mysql: [],
      s3: [],
      openai: [],
      mongodb: [],
      gcpcs: [],
      dynamodb: [],
    };
    expect(result).toMatchObject(expected);
  });
  it('should return an object with only postgres and mysql connections as a property', () => {
    const connections: Connection[] = [postgres, mysql];
    const result = splitConnections(connections);
    const expected = {
      postgres: [postgres],
      mysql: [mysql],
      s3: [],
      openai: [],
      mongodb: [],
      gcpcs: [],
      dynamodb: [],
    };
    expect(result).toMatchObject(expected);
  });
  it('should return an object with each connection as a property', () => {
    const connections: Connection[] = [
      postgres,
      mysql,
      mssql,
      s3,
      openai,
      mongodb,
      gcpcs,
      dynamodb,
    ];
    const result = splitConnections(connections);
    const expected = {
      postgres: [postgres],
      mysql: [mysql],
      mssql: [mssql],
      s3: [s3],
      openai: [openai],
      mongodb: [mongodb],
      gcpcs: [gcpcs],
      dynamodb: [dynamodb],
    };
    expect(result).toMatchObject(expected);
  });
  it('should return an object with keys but all empty arrays', () => {
    const connections: Connection[] = [];
    const result = splitConnections(connections);
    const expected = {
      postgres: [],
      mysql: [],
      mssql: [],
      s3: [],
      openai: [],
      mongodb: [],
      mssql: [],
      gcpcs: [],
      dynamodb: [],
    };
    expect(result).toMatchObject(expected);
  });
});

describe('refreshJobRunWhenJobRunning', () => {
  const TEN_SECONDS = 10 * 1000;

  it('should return 0 since COMPLETE is not in the list of status that shoudl return 10 seconds', () => {
    const data = new GetJobRunResponse({
      jobRun: new JobRun({
        status: JobRunStatus.COMPLETE,
      }),
    });
    const result = refreshJobRunWhenJobRunning(data);
    expect(result).toEqual(0);
  });
  it('should return 10 seconds in milliseconds since RUNNING is in th elist of status should return 10 seconds', () => {
    const data = new GetJobRunResponse({
      jobRun: new JobRun({
        status: JobRunStatus.RUNNING,
      }),
    });
    const result = refreshJobRunWhenJobRunning(data);
    expect(result).toEqual(TEN_SECONDS);
  });
  it('should return 10 seconds in milliseconds since PENDING is in th elist of status should return 10 seconds', () => {
    const data = new GetJobRunResponse({
      jobRun: new JobRun({
        status: JobRunStatus.PENDING,
      }),
    });
    const result = refreshJobRunWhenJobRunning(data);
    expect(result).toEqual(TEN_SECONDS);
  });
  it('should return 10 seconds in milliseconds since ERROR is in th elist of status should return 10 seconds', () => {
    const data = new GetJobRunResponse({
      jobRun: new JobRun({
        status: JobRunStatus.ERROR,
      }),
    });
    const result = refreshJobRunWhenJobRunning(data);
    expect(result).toEqual(TEN_SECONDS);
  });
  it('should return 0 since CANCELED is not in the list of status that shoudl return 10 seconds', () => {
    const data = new GetJobRunResponse({
      jobRun: new JobRun({
        status: JobRunStatus.CANCELED,
      }),
    });
    const result = refreshJobRunWhenJobRunning(data);
    expect(result).toEqual(0);
  });
  it('should return 0 because there is no jobRun.status', () => {
    const data = new GetJobRunResponse({
      jobRun: new JobRun({}),
    });
    const result = refreshJobRunWhenJobRunning(data);
    expect(result).toEqual(0);
  });
  it('should return 0 because there is no jobRun', () => {
    const data = new GetJobRunResponse({});
    const result = refreshJobRunWhenJobRunning(data);
    expect(result).toEqual(0);
  });
});

describe('refreshEventsWhenEventsIncomplete', () => {
  it('should return 0 since isRunComplete is true', () => {
    const input = new GetJobRunEventsResponse({
      events: [],
      isRunComplete: true,
    });
    const result = refreshEventsWhenEventsIncomplete(input);
    expect(result).toEqual(0);
  });
  it('should return 10 seconds in milliseconds since isRunComplete is false', () => {
    const input = new GetJobRunEventsResponse({
      events: [],
      isRunComplete: false,
    });
    const result = refreshEventsWhenEventsIncomplete(input);
    expect(result).toEqual(10 * 1000);
  });
  it('should return 10 seconds in milliseconds since isRunComplete is not there', () => {
    const input = new GetJobRunEventsResponse({
      events: [],
    });
    const result = refreshEventsWhenEventsIncomplete(input);
    expect(result).toEqual(10 * 1000);
  });
});

describe('onJobRunsAutoRefreshInterval', () => {
  it('should return off since interval is off', () => {
    const input: JobRunsAutoRefreshInterval = 'off';
    const result = onJobRunsAutoRefreshInterval(input);
    expect(result).toEqual(0);
  });
  it('should return 10000 since interval is 10s', () => {
    const input: JobRunsAutoRefreshInterval = '10s';
    const result = onJobRunsAutoRefreshInterval(input);
    expect(result).toEqual(10000);
  });
  it('should return 30000 since interval is 30s', () => {
    const input: JobRunsAutoRefreshInterval = '30s';
    const result = onJobRunsAutoRefreshInterval(input);
    expect(result).toEqual(30 * 1000);
  });
  it('should return 60000 in milliseconds since interval is 1m', () => {
    const input: JobRunsAutoRefreshInterval = '1m';
    const result = onJobRunsAutoRefreshInterval(input);
    expect(result).toEqual(1 * 60 * 1000);
  });
  it('should return 300000 since interval is 5m', () => {
    const input: JobRunsAutoRefreshInterval = '5m';
    const result = onJobRunsAutoRefreshInterval(input);
    expect(result).toBe(5 * 60 * 1000);
  });
  it('should return 0 since interval is invalid', () => {
    const input: JobRunsAutoRefreshInterval =
      'testest' as unknown as JobRunsAutoRefreshInterval;
    const result = onJobRunsAutoRefreshInterval(input);
    expect(result).toEqual(0);
  });
});

describe('onJobRunsPaused', () => {
  it('should return true since interval is off', () => {
    const input: JobRunsAutoRefreshInterval = 'off';
    const result = onJobRunsPaused(input);
    expect(result).toBe(true);
  });
  it('should return false since interval is 10s', () => {
    const input: JobRunsAutoRefreshInterval = '10s';
    const result = onJobRunsPaused(input);
    expect(result).toBe(false);
  });
  it('should return false since interval is 30s', () => {
    const input: JobRunsAutoRefreshInterval = '30s';
    const result = onJobRunsPaused(input);
    expect(result).toBe(false);
  });
  it('should return false since interval is 1m', () => {
    const input: JobRunsAutoRefreshInterval = '1m';
    const result = onJobRunsPaused(input);
    expect(result).toBe(false);
  });
  it('should return false since interval is 5m', () => {
    const input: JobRunsAutoRefreshInterval = '5m';
    const result = onJobRunsPaused(input);
    expect(result).toBe(false);
  });
  it('should return false since interval is invalid', () => {
    const input: JobRunsAutoRefreshInterval =
      'testest' as unknown as JobRunsAutoRefreshInterval;
    const result = onJobRunsPaused(input);
    expect(result).toBe(false);
  });
});
