'use client';

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
import {
  formatDateTime,
  getAccountRoleString,
  getErrorMessage,
} from '@/util/util';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { AccountRole, UserAccountService } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { useMemo, useState } from 'react';
import { toast } from 'sonner';
import { buildInviteLink } from './InviteUserForm';

interface MemberInviteRow {
  id: string;
  email: string;
  createdAt: Date;
  expiresAt: Date;
  token: string;
  role: AccountRole;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getColumns(isRbacEnabled: boolean): ColumnDef<MemberInviteRow, any>[] {
  const columnHelper = createColumnHelper<MemberInviteRow>();
  const emailColumn = columnHelper.accessor('email', {
    header: 'Email',
    cell: ({ getValue }) => <div>{getValue()}</div>,
  });

  const createdAtColumn = columnHelper.accessor('createdAt', {
    header: 'Created At',
    cell: ({ getValue }) => {
      return <div className="flex space-x-2">{formatDateTime(getValue())}</div>;
    },
  });

  const expiresAtColumn = columnHelper.accessor('expiresAt', {
    header: 'Expires At',
    cell: ({ getValue }) => {
      return <div className="flex space-x-2">{formatDateTime(getValue())}</div>;
    },
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
        <div className="flex flex-row gap-2">
          <CopyInviteButton token={row.original.token} />
          <DeleteInviteButton
            onDeleted={() =>
              table.options.meta?.invitesTable?.onDeleted(row.original.id)
            }
            inviteId={row.original.id}
          />
        </div>
      );
    },
  });

  if (isRbacEnabled) {
    return [
      emailColumn,
      createdAtColumn,
      expiresAtColumn,
      roleColumn,
      actionsColumn,
    ];
  }

  return [emailColumn, createdAtColumn, expiresAtColumn, actionsColumn];
}

function useGetColumns(): ColumnDef<MemberInviteRow>[] {
  const { data: config } = useGetSystemAppConfig();
  const isRbacEnabled = config?.isRbacEnabled ?? false;
  return useMemo(() => {
    return getColumns(isRbacEnabled);
  }, [isRbacEnabled]);
}

interface Props {
  accountId: string;
}

export function InvitesTable(props: Props): React.ReactElement {
  const { accountId } = props;
  const { data, isLoading, refetch, isFetching } = useQuery(
    UserAccountService.method.getTeamAccountInvites,
    { accountId: accountId },
    { enabled: !!accountId }
  );
  const invites = data?.invites || [];
  const invitesRows = useMemo(() => {
    return invites.map((invite): MemberInviteRow => {
      return {
        id: invite.id,
        email: invite.email,
        createdAt: invite.createdAt
          ? timestampDate(invite.createdAt)
          : new Date(),
        expiresAt: invite.expiresAt
          ? timestampDate(invite.expiresAt)
          : new Date(),
        token: invite.token,
        role: invite.role,
      };
    });
  }, [isFetching, invites]);

  const columns = useGetColumns();

  if (isLoading) {
    return <SkeletonTable />;
  }

  return (
    <DataTable
      data={invitesRows}
      columns={columns}
      onDeleted={() => refetch()}
    />
  );
}

declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface TableMeta<TData extends RowData> {
    invitesTable?: {
      onDeleted(id: string): void;
    };
  }
}

interface DataTableProps {
  data: MemberInviteRow[];
  columns: ColumnDef<MemberInviteRow>[];
  onDeleted(id: string): void;
}
function DataTable(props: DataTableProps): React.ReactElement {
  const { data, columns, onDeleted } = props;
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
      invitesTable: {
        onDeleted,
      },
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
  const { mutateAsync } = useMutation(
    UserAccountService.method.removeTeamAccountInvite
  );

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
