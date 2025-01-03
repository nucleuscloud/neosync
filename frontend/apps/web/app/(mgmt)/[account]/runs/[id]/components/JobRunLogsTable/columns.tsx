'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Timestamp, timestampDate } from '@bufbuild/protobuf/wkt';
import { GetJobRunLogsStreamResponse } from '@neosync/sdk';
import { DataTableColumnHeader } from './data-table-column-header';

interface GetColumnsProps {}

export function getColumns(
  props: GetColumnsProps
): ColumnDef<GetJobRunLogsStreamResponse>[] {
  const {} = props;
  return [
    {
      accessorKey: 'timestamp',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Timestamp" />
      ),
      size: 210,
      cell: ({ getValue, cell }) => {
        const date = getValue<Timestamp | undefined>();
        const text = date ? timestampDate(date).toISOString() : '-';
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
          <div className="flex space-x-2">
            <p className="font-medium text-wrap truncate">
              {getValue<string>()}
            </p>
          </div>
        );
      },
    },
  ];
}
