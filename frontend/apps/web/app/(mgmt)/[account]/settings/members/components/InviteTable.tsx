'use client';

import {
  ColumnDef,
  ColumnFiltersState,
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

import { CopyButton } from '@/components/CopyButton';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
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
import { formatDateTime, getErrorMessage } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { AccountInvite } from '@neosync/sdk';
import {
  getTeamAccountInvites,
  removeTeamAccountInvite,
} from '@neosync/sdk/connectquery';
import { TrashIcon } from '@radix-ui/react-icons';
import { toast } from 'sonner';
import { buildInviteLink } from './InviteUserForm';

interface ColumnProps {
  onDeleted(id: string): void;
}

function getColumns(
  props: ColumnProps
): ColumnDef<PlainMessage<AccountInvite>>[] {
  const { onDeleted } = props;
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
              {formatDateTime(row.getValue<Timestamp>('createdAt')?.toDate())}
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
              {formatDateTime(row.getValue<Timestamp>('expiresAt')?.toDate())}
            </span>
          </div>
        );
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => {
        return (
          <div className="flex flex-row gap-2">
            <CopyInviteButton token={row.original.token} />
            <DeleteInviteButton
              onDeleted={onDeleted}
              inviteId={row.original.id}
            />
          </div>
        );
      },
    },
  ];
}

interface Props {
  accountId: string;
}

export function InvitesTable(props: Props) {
  const { accountId } = props;
  const { data, isLoading, refetch } = useQuery(
    getTeamAccountInvites,
    { accountId: accountId },
    { enabled: !!accountId }
  );
  if (isLoading) {
    return <SkeletonTable />;
  }

  return <DataTable data={data?.invites || []} onDeleted={() => refetch()} />;
}

interface DataTableProps {
  data: AccountInvite[];
  onDeleted(id: string): void;
}
function DataTable(props: DataTableProps): React.ReactElement {
  const { data, onDeleted } = props;
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  const columns = React.useMemo(() => getColumns({ onDeleted }), []);

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
      <div className="flex items-center py-4 justify-between">
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

interface DeleteInviteButtonProps {
  inviteId: string;
  onDeleted(id: string): void;
}

function DeleteInviteButton({ inviteId, onDeleted }: DeleteInviteButtonProps) {
  const { mutateAsync } = useMutation(removeTeamAccountInvite);

  async function onRemove(): Promise<void> {
    try {
      await mutateAsync({ id: inviteId });
      toast.success('Invite deleted successfully!');
      onDeleted(inviteId);
    } catch (err) {
      console.error(err);
      toast.error('Unable to delete user invite', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <DeleteConfirmationDialog
      trigger={
        <Button variant="destructive" size="icon">
          <TrashIcon />
        </Button>
      }
      headerText="Are you sure you want to delete this invite?"
      description=""
      onConfirm={async () => onRemove()}
    />
  );
}

interface CopyInviteButtonProps {
  token: string;
}

function CopyInviteButton({ token }: CopyInviteButtonProps) {
  const { data: systemAppData } = useGetSystemAppConfig();
  const link = buildInviteLink(systemAppData?.publicAppBaseUrl ?? '', token);

  return (
    <CopyButton
      buttonVariant="outline"
      textToCopy={link}
      onCopiedText="Success!"
      onHoverText="Copy the invite link"
    />
  );
}
