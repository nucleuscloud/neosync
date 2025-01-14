'use client';
import { getDefaultUnmappedTransformConfig } from '@/app/(mgmt)/[account]/jobs/util';
import SwitchCard from '@/components/switches/SwitchCard';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { SourceOptionsFormValues } from '@/yup-validations/jobs';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import DynamoDBSourceOptionsForm from './DynamoDBSourceOptionsForm';
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
              columnRemovalStrategy: 'auto',
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
        <div className="flex flex-col gap-2">
          <div>
            <SwitchCard
              isChecked={value.mysql?.haltOnNewColumnAddition ?? false}
              onCheckedChange={(checked) => {
                setValue({
                  mysql: {
                    ...(value.mysql ?? { haltOnNewColumnAddition: false }),
                    haltOnNewColumnAddition: checked,
                  },
                });
              }}
              title="Halt Job on new column addition"
              description="Stops job runs if new column is detected"
            />
          </div>
        </div>
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
        <div className="flex flex-col gap-2">
          <div>
            <SwitchCard
              isChecked={value.mssql?.haltOnNewColumnAddition ?? false}
              onCheckedChange={(checked) => {
                setValue({
                  mysql: {
                    ...(value.mssql ?? {
                      haltOnNewColumnAddition: false,
                    }),
                    haltOnNewColumnAddition: checked,
                  },
                });
              }}
              title="Halt Job on new column addition"
              description="Stops job runs if new column is detected"
            />
          </div>
        </div>
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
