'use client';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
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
import { getErrorMessage } from '@/util/util';
import { PlainMessage } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { AccountUser } from '@neosync/sdk';
import {
  getTeamAccountMembers,
  removeTeamAccountMember,
} from '@neosync/sdk/connectquery';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import {
  ColumnDef,
  ColumnFiltersState,
  Row,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement, useMemo, useState } from 'react';
import { toast } from 'sonner';

interface ColumnProps {
  onDeleted(id: string): void;
  accountId: string;
}

function getColumns(
  props: ColumnProps
): ColumnDef<PlainMessage<AccountUser>>[] {
  const { onDeleted, accountId } = props;
  return [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => (
        <div className="flex flex-row items-center gap-4">
          <Avatar className="mr-2 h-12 w-12">
            <AvatarImage
              src={
                row.original.image ||
                `https://avatar.vercel.sh/${row.getValue('name')}.png`
              }
              alt={row.getValue('name')}
            />
          </Avatar>
          <span className=" truncate font-medium">{row.getValue('name')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'email',
      header: 'Email',
      cell: ({ row }) => (
        <span className=" truncate font-medium">{row.getValue('email')}</span>
      ),
    },
    {
      id: 'actions',
      cell: ({ row }) => (
        <DataTableRowActions
          accountId={accountId}
          row={row}
          onDeleted={() => onDeleted(row.id)}
        />
      ),
    },
  ];
}

interface Props {
  accountId: string;
}

export default function MembersTable(props: Props): ReactElement {
  const { accountId } = props;
  const { data, isLoading, refetch } = useQuery(
    getTeamAccountMembers,
    { accountId: accountId },
    { enabled: !!accountId }
  );
  const columns = useMemo(
    () => getColumns({ accountId, onDeleted: () => refetch() }),
    [accountId, refetch]
  );
  const users = data?.users || [];
  if (isLoading) {
    return <SkeletonTable />;
  }
  return <DataTable data={users} columns={columns} />;
}

interface DataTableProps {
  data: AccountUser[];
  columns: ColumnDef<PlainMessage<AccountUser>>[];
}

function DataTable(props: DataTableProps): React.ReactElement {
  const { data, columns } = props;
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

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
  accountId: string;
}

function DataTableRowActions<TData>({
  row,
  onDeleted,
  accountId,
}: DataTableRowActionsProps<TData>) {
  const user = row.original as AccountUser;
  const { mutateAsync } = useMutation(removeTeamAccountMember);

  async function onRemove(): Promise<void> {
    try {
      await mutateAsync({ accountId: accountId, userId: user.id });
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
