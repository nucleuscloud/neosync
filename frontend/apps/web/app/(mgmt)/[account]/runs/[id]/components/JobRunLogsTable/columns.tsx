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
        <DataTableColumnHeader column={column} title="Timestamp" />
      ),
      cell: ({ getValue, cell }) => {
        const date = getValue<Timestamp | undefined>()?.toDate();
        const text = date ? date.toISOString() : '-';
        return (
          <div
            className="flex space-x-2"
            style={{ maxWidth: cell.column.getSize() }}
          >
            <p className="font-medium">{text}</p>
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
          <p className="font-medium text-wrap truncate">{getValue<string>()}</p>
        );
      },
    },
  ];
}
