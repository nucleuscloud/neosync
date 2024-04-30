'use client';

import { ColumnDef } from '@tanstack/react-table';

import { PlainMessage } from '@bufbuild/protobuf';
import { ConnectionRolePrivilege } from '@neosync/sdk';
import { CheckCircledIcon, CircleBackslashIcon } from '@radix-ui/react-icons';
import { DataTableColumnHeader } from './data-table-column-header';

export function getPermissionColumns(): ColumnDef<
  PlainMessage<ConnectionRolePrivilege>
>[] {
  return [
    {
      accessorKey: 'Role',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Role" />
      ),
      cell: ({ row }) => <div>{row.original.grantee}</div>,
    },

    {
      accessorKey: 'schema',
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'table',
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorFn: (row) => `${row.schema}.${row.table}`,
      id: 'schemaTable',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Table" />
      ),
    },
    {
      id: 'read',
      accessorFn: (row) => row.privilegeType.includes('SELECT'),
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Read" />
      ),
      cell: ({ getValue }) => {
        return (
          <div>
            {getValue<boolean>() ? (
              <CheckCircledIcon className="text-green-500" />
            ) : (
              <CircleBackslashIcon className="text-red-500" />
            )}
          </div>
        );
      },
    },
    {
      id: 'create',
      accessorFn: (row) => row.privilegeType.includes('INSERT'),
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Create" />
      ),
      cell: ({ getValue }) => {
        return (
          <div>
            {getValue<boolean>() ? (
              <CheckCircledIcon className="text-green-500" />
            ) : (
              <CircleBackslashIcon className="text-red-500" />
            )}
          </div>
        );
      },
    },
    {
      id: 'update',
      accessorFn: (row) => row.privilegeType.includes('UPDATE'),
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Update" />
      ),
      cell: ({ getValue }) => {
        return (
          <div>
            {getValue<boolean>() ? (
              <CheckCircledIcon className="text-green-500" />
            ) : (
              <CircleBackslashIcon className="text-red-500" />
            )}
          </div>
        );
      },
    },
    {
      id: 'truncate',
      accessorFn: (row) => row.privilegeType.includes('TRUNCATE'),
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Truncate" />
      ),
      cell: ({ getValue }) => {
        return (
          <div>
            {getValue<boolean>() ? (
              <CheckCircledIcon className="text-green-500" />
            ) : (
              <CircleBackslashIcon className="text-red-500" />
            )}
          </div>
        );
      },
    },
  ];
}
