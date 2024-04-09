'use client';

import { ColumnDef } from '@tanstack/react-table';

import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { GetJobRunLogsStreamResponse } from '@neosync/sdk';
import { DataTableColumnHeader } from './data-table-column-header';

interface GetColumnsProps {}

export function getColumns(
  props: GetColumnsProps
): ColumnDef<PlainMessage<GetJobRunLogsStreamResponse>>[] {
  const {} = props;
  return [
    {
      accessorKey: 'timestamp',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Ts" />
      ),
      cell: ({ getValue }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {getValue<Timestamp | undefined>()?.toDate()?.toISOString() ??
                '-'}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'logLine',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Log" />
      ),
      cell: ({ getValue }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {getValue<string>()}
            </span>
          </div>
        );
      },
    },
  ];
}
