'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import { DestinationOptionsFormValues } from '@/yup-validations/jobs';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { FieldErrors } from 'react-hook-form';
import { DestinationDetails } from '../NosqlTable/TableMappings/Columns';
import TableMappingsCard, {
  Props as TableMappingsCardProps,
} from '../NosqlTable/TableMappings/TableMappingsCard';
import AwsS3DestinationOptionsForm from './AwsS3DestinationOptionsForm';
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
            }
          }
          setValue={(val) => setValue({ ...value, postgres: { ...val } })}
          hideInitTableSchema={hideInitTableSchema}
          errors={errors?.postgres}
        />
      );
    }

    case 'mysqlConfig':
      return (
        <div className="flex flex-col gap-2">
          <div>
            <SwitchCard
              isChecked={value.mysql?.truncateBeforeInsert ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mysql: {
                    ...(value.mysql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      skipForeignKeyViolations: false,
                    }),

                    truncateBeforeInsert: newVal,
                  },
                });
              }}
              title="Truncate Before Insert"
              description="Truncates table before inserting data"
            />
          </div>
          <div>
            <SwitchCard
              isChecked={value.mysql?.initTableSchema ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mysql: {
                    ...(value.mysql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      skipForeignKeyViolations: false,
                    }),
                    initTableSchema: newVal,
                  },
                });
              }}
              title="Init Table Schema"
              description="Creates table(s) and their constraints. The database schema must already exist. "
            />
          </div>
          <div>
            <SwitchCard
              isChecked={value.mysql?.onConflictDoNothing ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mysql: {
                    ...(value.mysql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      skipForeignKeyViolations: false,
                    }),
                    onConflictDoNothing: newVal,
                  },
                });
              }}
              title="On Conflict Do Nothing"
              description="If there is a conflict when inserting data do not insert"
            />
          </div>
          <div>
            <SwitchCard
              isChecked={value.mysql?.skipForeignKeyViolations ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mysql: {
                    ...(value.mysql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      truncateCascade: false,
                      skipForeignKeyViolations: false,
                    }),

                    skipForeignKeyViolations: newVal,
                  },
                });
              }}
              title="Skip Foreign Key Violations"
              description="Insert all valid records, bypassing any that violate foreign key constraints."
            />
          </div>
        </div>
      );
    case 'awsS3Config': {
      return (
        <AwsS3DestinationOptionsForm
          value={value.awss3 ?? {}}
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
        <div className="flex flex-col gap-2">
          <div>
            <SwitchCard
              isChecked={value.mssql?.truncateBeforeInsert ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mssql: {
                    ...(value.mssql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      skipForeignKeyViolations: false,
                    }),

                    truncateBeforeInsert: newVal,
                  },
                });
              }}
              title="Truncate Before Insert"
              description="Truncates table before inserting data"
            />
          </div>
          <div>
            {/* <SwitchCard
              isChecked={value.mssql?.initTableSchema ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mssql: {
                    ...(value.mssql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                    }),
                    initTableSchema: newVal,
                  },
                });
              }}
              title="Init Table Schema"
              description="Creates table(s) and their constraints. The database schema must already exist. "
            /> */}
          </div>
          <div>
            {/* <SwitchCard
              isChecked={value.mssql?.onConflictDoNothing ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mssql: {
                    ...(value.mssql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                    }),
                    onConflictDoNothing: newVal,
                  },
                });
              }}
              title="On Conflict Do Nothing"
              description="If there is a conflict when inserting data do not insert"
            /> */}
          </div>
          <div>
            <SwitchCard
              isChecked={value.mssql?.skipForeignKeyViolations ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  mssql: {
                    ...(value.mssql ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      truncateCascade: false,
                      skipForeignKeyViolations: false,
                    }),

                    skipForeignKeyViolations: newVal,
                  },
                });
              }}
              title="Skip Foreign Key Violations"
              description="Insert all valid records, bypassing any that violate foreign key constraints."
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
