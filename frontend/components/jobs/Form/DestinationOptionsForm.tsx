'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { ReactElement } from 'react';

interface DestinationOptionsProps {
  connection?: Connection;
  index: number;
  maxColNum?: number;
}
export default function DestinationOptionsForm(
  props: DestinationOptionsProps
): ReactElement {
  const { connection, maxColNum, index } = props;
  const grid = maxColNum ? `lg:grid-cols-${maxColNum}` : 'lg:grid-cols-3';

  if (!connection) {
    return <></>;
  }
  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      const value = connection.connectionConfig.config.value;
      switch (value.connectionConfig.case) {
        case 'connection':
          return (
            <div className={`grid grid-cols-1 md:grid-cols-1 ${grid} gap-4`}>
              <div>
                <FormField
                  name={`destinations.${index}.destinationOptions.truncateBeforeInsert`}
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <SwitchCard
                          isChecked={field.value || false}
                          onCheckedChange={field.onChange}
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
                  name={`destinations.${index}.destinationOptions.initDbSchema`}
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <SwitchCard
                          isChecked={field.value || false}
                          onCheckedChange={field.onChange}
                          title="Init Database Schema"
                          description="Creates database schema"
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
