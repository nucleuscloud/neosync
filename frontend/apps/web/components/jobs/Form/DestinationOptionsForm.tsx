'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import { Badge } from '@/components/ui/badge';
import { DestinationOptionsFormValues } from '@/yup-validations/jobs';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';

interface DestinationOptionsProps {
  connection?: Connection;

  value: DestinationOptionsFormValues;
  setValue(newVal: DestinationOptionsFormValues): void;

  hideInitTableSchema?: boolean;
  hideDynamoDbTableMappings?: boolean;
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
  } = props;

  if (!connection) {
    return <></>;
  }

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      return (
        <div className="flex flex-col gap-2">
          <div>
            <SwitchCard
              isChecked={value.postgres?.truncateBeforeInsert ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  postgres: {
                    ...(value.postgres ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      truncateCascade: false,
                    }),

                    truncateBeforeInsert: newVal,
                    truncateCascade: newVal
                      ? (value.postgres?.truncateCascade ?? false)
                      : false,
                  },
                });
              }}
              title="Truncate Before Insert"
              description="Truncates table before inserting data"
            />
          </div>
          <div>
            <SwitchCard
              isChecked={value.postgres?.truncateCascade ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  postgres: {
                    ...(value.postgres ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      truncateCascade: false,
                    }),

                    truncateBeforeInsert:
                      newVal && !value.postgres?.truncateBeforeInsert
                        ? true
                        : (value.postgres?.truncateBeforeInsert ?? false),
                    truncateCascade: newVal,
                  },
                });
              }}
              title="Truncate Cascade"
              description="TRUNCATE CASCADE to all tables"
            />
          </div>
          {!hideInitTableSchema && (
            <div>
              <SwitchCard
                isChecked={value.postgres?.initTableSchema ?? false}
                onCheckedChange={(newVal) => {
                  setValue({
                    ...value,
                    postgres: {
                      ...(value.postgres ?? {
                        initTableSchema: false,
                        onConflictDoNothing: false,
                        truncateBeforeInsert: false,
                        truncateCascade: false,
                      }),

                      initTableSchema: newVal ?? false,
                    },
                  });
                }}
                title="Init Table Schema"
                postTitle={<Badge>Experimental</Badge>}
                description="Creates table(s) and their constraints. The database schema must already exist. "
              />
            </div>
          )}
          <div>
            <SwitchCard
              isChecked={value.postgres?.onConflictDoNothing ?? false}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  postgres: {
                    ...(value.postgres ?? {
                      initTableSchema: false,
                      onConflictDoNothing: false,
                      truncateBeforeInsert: false,
                      truncateCascade: false,
                    }),

                    onConflictDoNothing: newVal,
                  },
                });
              }}
              title="On Conflict Do Nothing"
              description="If there is a conflict when inserting data do not insert"
            />
          </div>
        </div>
      );
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
                    }),
                    onConflictDoNothing: newVal,
                  },
                });
              }}
              title="On Conflict Do Nothing"
              description="If there is a conflict when inserting data do not insert"
            />
          </div>
        </div>
      );
    case 'awsS3Config':
      return <></>;
    case 'mongoConfig':
      return <></>;
    case 'gcpCloudstorageConfig':
      return <></>;
    case 'dynamodbConfig':
      return <></>;
    default:
      return (
        <div>
          No connection component found for: (
          {connection?.name ?? 'unknown name'})
        </div>
      );
  }
}
