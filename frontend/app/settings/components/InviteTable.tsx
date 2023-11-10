'use client';

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
import * as React from 'react';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
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
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountInvites } from '@/libs/hooks/useGetAccountInvites';
import { AccountInvite } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { formatDateTime, getErrorMessage } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';

interface ColumnProps {
  onDeleted(id: string): void;
  accountId: string;
}

function getColumns(
  props: ColumnProps
): ColumnDef<PlainMessage<AccountInvite>>[] {
  const { onDeleted, accountId } = props;
  return [
    {
      accessorKey: 'email',
      header: 'Email',
      cell: ({ row }) => <div>{row.getValue('email')}</div>,
    },
    {
      accessorKey: 'createdAt',
      header: 'Created At',
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(row.getValue<Timestamp>('createdAt').toDate())}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'expiresAt',
      header: 'Expires At',
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(row.getValue<Timestamp>('expiresAt').toDate())}
            </span>
          </div>
        );
      },
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

export function InvitesTable(props: Props) {
  const { accountId } = props;
  const { data, isLoading, mutate } = useGetAccountInvites(accountId);
  if (isLoading) {
    return <SkeletonTable />;
  }

  return (
    <DataTable
      data={data?.invites || []}
      accountId={accountId}
      onDeleted={() => mutate()}
    />
  );
}

interface DataTableProps {
  data: AccountInvite[];
  accountId: string;
  onDeleted(id: string): void;
}
function DataTable(props: DataTableProps): React.ReactElement {
  const { data, accountId, onDeleted } = props;
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  const columns = getColumns({ accountId, onDeleted });

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
    </div>
  );
}

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
  accountId: string;
}

export function DataTableRowActions<TData>({
  row,
  onDeleted,
  accountId,
}: DataTableRowActionsProps<TData>) {
  const invite = row.original as AccountInvite;
  const { toast } = useToast();

  async function onRemove(): Promise<void> {
    try {
      await deleteAccountInvite(accountId, invite.id);
      toast({
        title: 'Invited deleted successfully!',
      });
      onDeleted();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to delete invite',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <DropdownMenu
      modal={false} // needed because otherwise this breaks after a single use in conjunction with the delete dialog
    >
      <DropdownMenuTrigger className="hover:bg-gray-100 py-1 px-2 rounded-lg">
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
          headerText="Are you sure you want to delete this invite?"
          description=""
          onConfirm={() => onRemove()}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

async function deleteAccountInvite(
  accountId: string,
  inviteId: string
): Promise<void> {
  const res = await fetch(
    `/api/users/accounts/${accountId}/invites?id=${inviteId}`,
    {
      method: 'DELETE',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
