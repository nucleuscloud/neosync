'use client';

import { ColumnDef } from '@tanstack/react-table';

import { SchemaColumnHeader } from '@/components/jobs/SchemaTable/SchemaColumnHeader';
import TruncatedText from '@/components/TruncatedText';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { GetJobRunLogsResponse_LogLine } from '@neosync/sdk';

export function getColumns(): ColumnDef<GetJobRunLogsResponse_LogLine>[] {
  return [
    {
      id: 'timestamp',
      accessorFn: (row) => {
        const date = row.timestamp;
        return date ? timestampDate(date).toISOString() : '-';
      },
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Timestamp" />
      ),
      size: 210,
      cell: ({ getValue, cell }) => {
        return (
          <TruncatedText
            text={getValue<string>()}
            maxWidth={cell.column.getSize()}
          />
        );
      },
    },
    {
      accessorKey: 'logLine',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Log" />
      ),
      cell: ({ getValue, cell }) => {
        return (
          <TruncatedText
            text={getValue<string>()}
            maxWidth={cell.column.getSize()}
          />
        );
      },
      size: 1000,
    },
    {
      id: 'labels',
      accessorFn: (row) => {
        const value = row.labels ?? {};
        return JSON.stringify(value);
      },
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Labels" />
      ),
      cell: ({ getValue, cell }) => {
        return (
          <TruncatedText
            text={getValue<string>()}
            maxWidth={cell.column.getSize()}
          />
        );
      },
      size: 500,
    },
  ];
}
