'use client';
import { getDefaultUnmappedTransformConfig } from '@/app/(mgmt)/[account]/jobs/util';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { SourceOptionsFormValues } from '@/yup-validations/jobs';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import DynamoDBSourceOptionsForm from './DynamoDBSourceOptionsForm';
import MssqlDBSourceOptionsForm from './MssqlDBSourceOptionsForm';
import MysqlDBSourceOptionsForm from './MysqlDBSourceOptionsForm';
import PostgresDBSourceOptionsForm from './PostgresDBSourceOptionsForm';

interface SourceOptionsProps {
  connection?: Connection;
  value: SourceOptionsFormValues;
  setValue(newVal: SourceOptionsFormValues): void;
}
export default function SourceOptionsForm(
  props: SourceOptionsProps
): ReactElement {
  const { connection, value, setValue } = props;

  const {
    handler: transformersHandler,
    isLoading: isTransformersHandlerLoading,
  } = useGetTransformersHandler(connection?.accountId ?? '');

  if (!connection || isTransformersHandlerLoading) {
    return <></>;
  }

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      return (
        <PostgresDBSourceOptionsForm
          value={
            value.postgres ?? {
              newColumnAdditionStrategy: 'continue',
              columnRemovalStrategy: 'continue',
            }
          }
          setValue={(newval) => {
            setValue({
              postgres: {
                ...(value.postgres ?? {}),
                ...newval,
              },
            });
          }}
        />
      );
    case 'mysqlConfig':
      return (
        <MysqlDBSourceOptionsForm
          value={
            value.mysql ?? {
              haltOnNewColumnAddition: false,
              columnRemovalStrategy: 'continue',
            }
          }
          setValue={(newval) => {
            setValue({
              mysql: {
                ...(value.mysql ?? {}),
                ...newval,
              },
            });
          }}
        />
      );
    case 'awsS3Config':
      return <></>;
    case 'openaiConfig':
      return <></>;
    case 'mongoConfig':
      return <></>;
    case 'gcpCloudstorageConfig':
      return <></>;
    case 'dynamodbConfig':
      return (
        <DynamoDBSourceOptionsForm
          value={
            value.dynamodb ?? {
              unmappedTransformConfig: getDefaultUnmappedTransformConfig(),
              enableConsistentRead: false,
            }
          }
          setValue={(newVal) => {
            setValue({
              dynamodb: {
                ...(value.dynamodb ?? {}),
                ...newVal,
              },
            });
          }}
          transformerHandler={transformersHandler}
        />
      );
    case 'mssqlConfig': {
      return (
        <MssqlDBSourceOptionsForm
          value={
            value.mssql ?? {
              haltOnNewColumnAddition: false,
              columnRemovalStrategy: 'continue',
            }
          }
          setValue={(newval) => {
            setValue({
              mssql: {
                ...(value.mssql ?? {}),
                ...newval,
              },
            });
          }}
        />
      );
    }
    default:
      return (
        <div>
          No connection component found for: (
          {connection?.name ?? 'unknown name'})
        </div>
      );
  }
}
