'use client';

import { ColumnDef } from '@tanstack/react-table';

import { ConnectionRolePrivilege } from '@neosync/sdk';
import { CheckCircledIcon, CircleBackslashIcon } from '@radix-ui/react-icons';
import { DataTableColumnHeader } from './data-table-column-header';

export type PermissionConnectionType =
  | 'mongodb'
  | 'mysql'
  | 'postgres'
  | 'dynamodb'
  | 'mssql';

export function getPermissionColumns(
  connectionType: PermissionConnectionType
): ColumnDef<ConnectionRolePrivilege>[] {
  switch (connectionType) {
    case 'mongodb':
    case 'dynamodb':
      return [
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
          accessorFn: (row) => {
            if (row.schema) {
              return `${row.schema}.${row.table}`;
            }
            return row.table;
          },
          id: 'schemaTable',
          header: ({ column }) => (
            <DataTableColumnHeader column={column} title="Collection" />
          ),
          cell: ({ getValue }) => {
            return <p className="truncate">{getValue<string>()}</p>;
          },
        },
      ];
    default:
      return [
        {
          accessorKey: 'grantee',
          header: ({ column }) => (
            <DataTableColumnHeader column={column} title="Role" />
          ),
          cell: ({ getValue }) => {
            return <p className="truncate">{getValue<string>()}</p>;
          },
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
          cell: ({ getValue }) => {
            return <p className="truncate">{getValue<string>()}</p>;
          },
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
          accessorFn: (row) =>
            row.privilegeType.includes('TRUNCATE') ||
            row.privilegeType.includes('DELETE'),
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
}
