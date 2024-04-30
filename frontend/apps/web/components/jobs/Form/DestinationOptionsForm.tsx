'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface DestinationOptionsProps {
  connection?: Connection;
  index?: number;
}
export default function DestinationOptionsForm(
  props: DestinationOptionsProps
): ReactElement {
  const { connection, index } = props;
  const formCtx = useFormContext();

  if (!connection) {
    return <></>;
  }

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      const truncateBeforeInsertName =
        index != null
          ? `destinations.${index}.destinationOptions.truncateBeforeInsert`
          : `destinationOptions.truncateBeforeInsert`;
      const truncateCascadeName =
        index != null
          ? `destinations.${index}.destinationOptions.truncateCascade`
          : `destinationOptions.truncateCascade`;
      return (
        <div className="flex flex-col gap-2">
          <div>
            <FormField
              name={truncateBeforeInsertName}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <SwitchCard
                      isChecked={field.value || false}
                      onCheckedChange={(newVal) => {
                        field.onChange(newVal);
                        if (!newVal) {
                          formCtx.setValue(truncateCascadeName, false);
                        }
                      }}
                      title="Truncate Before Insert"
                      description="Truncates table before inserting data"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          <div>
            <FormField
              name={truncateCascadeName}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <SwitchCard
                      isChecked={field.value || false}
                      onCheckedChange={(newVal) => {
                        field.onChange(newVal);
                        if (newVal) {
                          formCtx.setValue(truncateBeforeInsertName, true);
                        }
                      }}
                      title="Truncate Cascade"
                      description="TRUNCATE CASCADE to all tables"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          <div>
            <FormField
              name={
                index != null
                  ? `destinations.${index}.destinationOptions.initTableSchema`
                  : `destinationOptions.initTableSchema`
              }
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <SwitchCard
                      isChecked={field.value || false}
                      onCheckedChange={field.onChange}
                      title="Init Table Schema"
                      description="Creates table(s) and their constraints. The database schema must already exist. "
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          <div>
            <FormField
              name={
                index != null
                  ? `destinations.${index}.destinationOptions.onConflictDoNothing`
                  : `destinationOptions.onConflictDoNothing`
              }
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <SwitchCard
                      isChecked={field.value || false}
                      onCheckedChange={field.onChange}
                      title="On Conflict Do Nothing"
                      description="If there is a conflict when inserting data do not insert"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>
      );
    case 'mysqlConfig':
      const mysqlValue = connection.connectionConfig.config.value;
      const mysqltruncateBeforeInsertName =
        index != null
          ? `destinations.${index}.destinationOptions.truncateBeforeInsert`
          : `destinationOptions.truncateBeforeInsert`;
      switch (mysqlValue.connectionConfig.case) {
        case 'connection':
          return (
            <div className="flex flex-col gap-2">
              <div>
                <FormField
                  name={mysqltruncateBeforeInsertName}
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <SwitchCard
                          isChecked={field.value || false}
                          onCheckedChange={(newVal) => {
                            field.onChange(newVal);
                          }}
                          title="Truncate Before Insert"
                          description="Truncates table before inserting data"
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <div>
                <FormField
                  name={
                    index != null
                      ? `destinations.${index}.destinationOptions.initTableSchema`
                      : `destinationOptions.initTableSchema`
                  }
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <SwitchCard
                          isChecked={field.value || false}
                          onCheckedChange={field.onChange}
                          title="Init Table Schema"
                          description="Creates table(s) and their constraints. The database schema must already exist. "
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <div>
                <FormField
                  name={
                    index != null
                      ? `destinations.${index}.destinationOptions.onConflictDoNothing`
                      : `destinationOptions.onConflictDoNothing`
                  }
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <SwitchCard
                          isChecked={field.value || false}
                          onCheckedChange={field.onChange}
                          title="On Conflict Do Nothing"
                          description="If there is a conflict when inserting data do not insert"
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </div>
          );
      }
      return <></>;
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
