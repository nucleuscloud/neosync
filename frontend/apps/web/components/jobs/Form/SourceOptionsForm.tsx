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

interface SourceOptionsProps {
  connection?: Connection;
}
export default function SourceOptionsForm(
  props: SourceOptionsProps
): ReactElement {
  const { connection } = props;

  if (!connection) {
    return <></>;
  }
  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      return (
        <div className="flex flex-col gap-2">
          <div>
            <FormField
              name="sourceOptions.haltOnNewColumnAddition"
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <SwitchCard
                      isChecked={field.value || false}
                      onCheckedChange={field.onChange}
                      title="Halt Job on new column addition"
                      description="Stops job runs if new column is detected"
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
      switch (mysqlValue.connectionConfig.case) {
        case 'connection':
          return (
            <div className="flex flex-col gap-2">
              <div>
                <FormField
                  name="sourceOptions.haltOnNewColumnAddition"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <SwitchCard
                          isChecked={field.value || false}
                          onCheckedChange={field.onChange}
                          title="Halt Job on new column addition"
                          description="Stops job runs if new column is detected"
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
    case 'openaiConfig':
      return <></>;
    case 'mongoConfig':
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
