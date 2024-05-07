'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import { Badge } from '@/components/ui/badge';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';

export interface DestinationOptions {
  truncateBeforeInsert: boolean;
  truncateCascade: boolean;
  initTableSchema: boolean;
  onConflictDoNothing: boolean;
}

interface DestinationOptionsProps {
  connection?: Connection;

  value: DestinationOptions;
  setValue(newVal: DestinationOptions): void;

  hideInitTableSchema?: boolean;
}

export default function DestinationOptionsForm(
  props: DestinationOptionsProps
): ReactElement {
  const { connection, value, setValue, hideInitTableSchema } = props;

  if (!connection) {
    return <></>;
  }

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      return (
        <div className="flex flex-col gap-2">
          <div>
            <SwitchCard
              isChecked={value.truncateBeforeInsert}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  truncateBeforeInsert: newVal,
                  truncateCascade: newVal ? value.truncateCascade : false,
                });
              }}
              title="Truncate Before Insert"
              description="Truncates table before inserting data"
            />
          </div>
          <div>
            <SwitchCard
              isChecked={value.truncateCascade}
              onCheckedChange={(newVal) => {
                setValue({
                  ...value,
                  truncateBeforeInsert:
                    newVal && !value.truncateBeforeInsert
                      ? true
                      : value.truncateBeforeInsert,
                  truncateCascade: newVal,
                });
              }}
              title="Truncate Cascade"
              description="TRUNCATE CASCADE to all tables"
            />
          </div>
          {!hideInitTableSchema && (
            <div>
              <SwitchCard
                isChecked={value.initTableSchema}
                onCheckedChange={(newVal) => {
                  setValue({ ...value, initTableSchema: newVal });
                }}
                title="Init Table Schema"
                postTitle={<Badge>Experimental</Badge>}
                description="Creates table(s) and their constraints. The database schema must already exist. "
              />
            </div>
          )}
          <div>
            <SwitchCard
              isChecked={value.onConflictDoNothing}
              onCheckedChange={(newVal) => {
                setValue({ ...value, onConflictDoNothing: newVal });
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
              isChecked={value.truncateBeforeInsert}
              onCheckedChange={(newVal) => {
                setValue({ ...value, truncateBeforeInsert: newVal });
              }}
              title="Truncate Before Insert"
              description="Truncates table before inserting data"
            />
          </div>
          <div>
            <SwitchCard
              isChecked={value.initTableSchema}
              onCheckedChange={(newVal) => {
                setValue({ ...value, initTableSchema: newVal });
              }}
              title="Init Table Schema"
              description="Creates table(s) and their constraints. The database schema must already exist. "
            />
          </div>
          <div>
            <SwitchCard
              isChecked={value.onConflictDoNothing}
              onCheckedChange={(newVal) => {
                setValue({ ...value, onConflictDoNothing: newVal });
              }}
              title="On Conflict Do Nothing"
              description="If there is a conflict when inserting data do not insert"
            />
          </div>
        </div>
      );
    case 'awsS3Config':
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
