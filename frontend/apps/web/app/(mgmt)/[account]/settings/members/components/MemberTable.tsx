'use client';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Avatar, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getAccountRoleString, getErrorMessage } from '@/util/util';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { AccountRole, AccountUser, UserAccountService } from '@neosync/sdk';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import {
  ColumnDef,
  ColumnFiltersState,
  RowData,
  SortingState,
  VisibilityState,
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement, useMemo, useState } from 'react';
import { toast } from 'sonner';
import UpdateMemberRoleDialog from './UpdateMemberRoleDialog';

interface MemberRow {
  id: string;
  name: string;
  email: string;
  image?: string;
  role: AccountRole;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getColumns(isRbacEnabled: boolean): ColumnDef<MemberRow, any>[] {
  const columnHelper = createColumnHelper<MemberRow>();
  const nameColumn = columnHelper.accessor('name', {
    header: 'Name',
    cell: ({ row, getValue }) => (
      <div className="flex flex-row items-center gap-4">
        <Avatar className="mr-2 h-12 w-12">
          <AvatarImage
            src={
              row.original.image || `https://avatar.vercel.sh/${getValue()}.png`
            }
            alt={getValue()}
          />
        </Avatar>
        <span className="truncate font-medium">{getValue()}</span>
      </div>
    ),
  });

  const emailColumn = columnHelper.accessor('email', {
    header: 'Email',
    cell: ({ getValue }) => (
      <span className="truncate font-medium">{getValue()}</span>
    ),
  });

  const roleColumn = columnHelper.accessor('role', {
    header: 'Role',
    cell: ({ getValue }) => (
      <span className="truncate font-medium">
        {getAccountRoleString(getValue())}
      </span>
    ),
  });

  const actionsColumn = columnHelper.display({
    id: 'actions',
    cell: ({ row, table }) => {
      return (
        <DataTableRowActions
          member={{
            id: row.original.id,
            name: row.original.name,
            role: row.original.role,
            email: row.original.email,
          }}
          onDeleted={() =>
            table.options.meta?.membersTable?.onDeleted(row.original.id)
          }
          onUpdated={() =>
            table.options.meta?.membersTable?.onUpdated(row.original.id)
          }
        />
      );
    },
  });

  if (isRbacEnabled) {
    return [nameColumn, emailColumn, roleColumn, actionsColumn];
  }

  return [nameColumn, emailColumn, actionsColumn];
}

interface Props {
  accountId: string;
}

export default function MembersTable(props: Props): ReactElement<any> {
  const { accountId } = props;
  const { data, isLoading, refetch, isFetching } = useQuery(
    UserAccountService.method.getTeamAccountMembers,
    { accountId: accountId },
    { enabled: !!accountId }
  );

  const users = data?.users ?? [];
  const members = useMemo(() => {
    return users.map((d): MemberRow => {
      return {
        name: d.name,
        email: d.email,
        image: d.image,
        id: d.id,
        role: d.role,
      };
    });
  }, [isFetching, users]);

  const columns = useGetColumns();

  if (isLoading) {
    return <SkeletonTable />;
  }
  return (
    <DataTable
      data={members}
      columns={columns}
      onDeleted={() => refetch()}
      onUpdated={() => refetch()}
    />
  );
}

function useGetColumns(): ColumnDef<MemberRow>[] {
  const { data: config } = useGetSystemAppConfig();
  const isRbacEnabled = config?.isRbacEnabled ?? false;
  return useMemo(() => {
    return getColumns(isRbacEnabled);
  }, [isRbacEnabled]);
}

declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface TableMeta<TData extends RowData> {
    membersTable?: {
      onDeleted(userId: string): void;
      onUpdated(userId: string): void;
    };
  }
}

interface DataTableProps {
  data: MemberRow[];
  columns: ColumnDef<MemberRow>[];
  onDeleted(userId: string): void;
  onUpdated(userId: string): void;
}

function DataTable(props: DataTableProps): React.ReactElement<any> {
  const { data, columns, onDeleted, onUpdated } = props;
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
    meta: {
      membersTable: {
        onDeleted,
        onUpdated,
      },
    },
  });

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        <Input
          placeholder="Filter emails..."
          value={(table.getColumn('email')?.getFilterValue() as string) ?? ''}
          onChange={(event) =>
            table.getColumn('email')?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}

interface DataTableRowActionsProps {
  member: Pick<AccountUser, 'id' | 'name' | 'role' | 'email'>;
  onDeleted(): void;
  onUpdated(): void;
}

function DataTableRowActions({
  member,
  onDeleted,
  onUpdated,
}: DataTableRowActionsProps) {
  const { account } = useAccount();

  const { mutateAsync } = useMutation(
    UserAccountService.method.removeTeamAccountMember
  );

  async function onRemove(): Promise<void> {
    if (!account?.id) {
      return;
    }
    try {
      await mutateAsync({ accountId: account.id, userId: member.id });
      toast.success('User removed from account!');
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove user from account!', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <DropdownMenu
      modal={false} // needed because otherwise this breaks after a single use in conjunction with the delete dialog
    >
      <DropdownMenuTrigger className="hover:bg-gray-100 dark:hover:bg-gray-800 py-1 px-2 rounded-lg">
        <DotsHorizontalIcon className="h-4 w-4" />
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <UpdateMemberRoleDialog
          member={member}
          onUpdated={() => onUpdated()}
          dialogButton={
            <DropdownMenuItem
              className="cursor-pointer"
              onSelect={(e) => e.preventDefault()}
            >
              Update Role
            </DropdownMenuItem>
          }
        />
        <DeleteConfirmationDialog
          trigger={
            <DropdownMenuItem
              className="cursor-pointer"
              onSelect={(e) => e.preventDefault()} // needed for the delete modal to not automatically close
            >
              Remove
            </DropdownMenuItem>
          }
          headerText="Are you sure you want to remove this user?"
          description=""
          onConfirm={() => onRemove()}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
