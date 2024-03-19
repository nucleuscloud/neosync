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
  maxColNum?: number;
}
export default function SourceOptionsForm(
  props: SourceOptionsProps
): ReactElement {
  const { connection, maxColNum } = props;
  const grid = maxColNum ? `lg:grid-cols-${maxColNum}` : 'lg:grid-cols-3';

  if (!connection) {
    return <></>;
  }
  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      const value = connection.connectionConfig.config.value;
      return (
        <div className={`grid grid-cols-1 md:grid-cols-1 ${grid} gap-4`}>
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
            <div className={`grid grid-cols-1 md:grid-cols-1 ${grid} gap-4`}>
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
    default:
      return (
        <div>
          No connection component found for: (
          {connection?.name ?? 'unknown name'})
        </div>
      );
  }
}
