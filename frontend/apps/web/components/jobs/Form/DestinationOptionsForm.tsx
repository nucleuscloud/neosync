'use client';
import { DestinationOptionsFormValues } from '@/yup-validations/jobs';
import {
  AwsS3DestinationConnectionOptions_StorageClass,
  Connection,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { FieldErrors } from 'react-hook-form';
import { DestinationDetails } from '../NosqlTable/TableMappings/Columns';
import TableMappingsCard, {
  Props as TableMappingsCardProps,
} from '../NosqlTable/TableMappings/TableMappingsCard';
import AwsS3DestinationOptionsForm from './AwsS3DestinationOptionsForm';
import MssqlDBDestinationOptionsForm from './MssqlDBDestinationOptionsForm';
import MysqlDBDestinationOptionsForm from './MysqlDBDestinationOptionsForm';
import PostgresDBDestinationOptionsForm from './PostgresDBDestinationOptionsForm';

interface DestinationOptionsProps {
  connection?: Connection;

  value: DestinationOptionsFormValues;
  setValue(newVal: DestinationOptionsFormValues): void;

  hideInitTableSchema?: boolean;
  hideDynamoDbTableMappings?: boolean;
  destinationDetailsRecord: Record<string, DestinationDetails>;
  errors?: FieldErrors<DestinationOptionsFormValues>;
}

export default function DestinationOptionsForm(
  props: DestinationOptionsProps
): ReactElement {
  const {
    connection,
    value,
    setValue,
    hideInitTableSchema,
    hideDynamoDbTableMappings,
    destinationDetailsRecord,
    errors,
  } = props;

  if (!connection) {
    return <></>;
  }

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig': {
      return (
        <PostgresDBDestinationOptionsForm
          value={
            value.postgres ?? {
              initTableSchema: false,
              onConflictDoNothing: false,
              skipForeignKeyViolations: false,
              truncateBeforeInsert: false,
              truncateCascade: false,
              batch: {
                count: 100,
                period: '5s',
              },
              maxInFlight: 64,
            }
          }
          setValue={(val) => setValue({ ...value, postgres: { ...val } })}
          hideInitTableSchema={hideInitTableSchema}
          errors={errors?.postgres}
        />
      );
    }

    case 'mysqlConfig': {
      return (
        <MysqlDBDestinationOptionsForm
          value={
            value.mysql ?? {
              initTableSchema: false,
              onConflictDoNothing: false,
              skipForeignKeyViolations: false,
              truncateBeforeInsert: false,
              batch: {
                count: 100,
                period: '5s',
              },
              maxInFlight: 64,
            }
          }
          setValue={(val) => setValue({ ...value, mysql: { ...val } })}
          hideInitTableSchema={hideInitTableSchema}
          errors={errors?.mysql}
        />
      );
    }

    case 'awsS3Config': {
      return (
        <AwsS3DestinationOptionsForm
          value={
            value.awss3 ?? {
              batch: {
                count: 100,
                period: '5s',
              },
              maxInFlight: 64,
              storageClass:
                AwsS3DestinationConnectionOptions_StorageClass.STANDARD,
              timeout: '5s',
            }
          }
          setValue={(val) => setValue({ ...value, awss3: { ...val } })}
          errors={errors?.awss3}
        />
      );
    }
    case 'mongoConfig':
      return <></>;
    case 'gcpCloudstorageConfig':
      return <></>;
    case 'dynamodbConfig':
      return (
        <DynamoDbOptions
          hideDynamoDbTableMappings={hideDynamoDbTableMappings ?? false}
          tableMappingsProps={{
            destinationDetailsRecord,
            mappings: value.dynamodb?.tableMappings
              ? [
                  {
                    destinationId: connection.id ?? '0',
                    dynamodb: { tableMappings: value.dynamodb.tableMappings },
                  },
                ]
              : [],
            onUpdate(req) {
              const tableMappings = value.dynamodb?.tableMappings ?? [];
              if (tableMappings.length === 0) {
                tableMappings.push({
                  sourceTable: req.souceName,
                  destinationTable: req.tableName,
                });
              } else {
                tableMappings.forEach((tm) => {
                  if (tm.sourceTable === req.souceName) {
                    tm.destinationTable = req.tableName;
                  }
                });
              }
              setValue({
                ...value,
                dynamodb: {
                  ...value.dynamodb,
                  tableMappings: tableMappings,
                },
              });
            },
          }}
        />
      );
    case 'mssqlConfig': {
      return (
        <MssqlDBDestinationOptionsForm
          value={
            value.mssql ?? {
              initTableSchema: false,
              onConflictDoNothing: false,
              skipForeignKeyViolations: false,
              truncateBeforeInsert: false,
              batch: {
                count: 100,
                period: '5s',
              },
              maxInFlight: 64,
            }
          }
          setValue={(val) => setValue({ ...value, mssql: { ...val } })}
          hideInitTableSchema={hideInitTableSchema}
          errors={errors?.mssql}
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

interface DynamoDbOptionsProps {
  hideDynamoDbTableMappings: boolean;

  tableMappingsProps: TableMappingsCardProps;
}

function DynamoDbOptions(props: DynamoDbOptionsProps): ReactElement {
  const { hideDynamoDbTableMappings, tableMappingsProps } = props;

  return (
    <div className="flex flex-col gap-2">
      {!hideDynamoDbTableMappings && (
        <TableMappingsCard {...tableMappingsProps} />
      )}
    </div>
  );
}
