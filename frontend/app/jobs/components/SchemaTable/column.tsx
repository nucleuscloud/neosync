'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Checkbox } from '@/components/ui/checkbox';

import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { DatabaseColumn } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { PlainMessage } from '@bufbuild/protobuf';
import { DataTableColumnHeader } from './data-table-column-header';

interface GetColumnsProps {
  transformers?: Transformer[];
}

export function getColumns(
  props: GetColumnsProps
): ColumnDef<PlainMessage<DatabaseColumn>>[] {
  const { transformers } = props;

  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
          className="translate-y-[2px]"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'schema',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Schema" />
      ),
      cell: ({ row }) => (
        <div className="w-[80px]">{row.getValue('schema')}</div>
      ),
      enableSorting: true,
    },
    {
      accessorKey: 'table',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Table" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('table')}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'column',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Column" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('column')}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'dataType',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Data Type" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('dataType')}
            </span>
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      accessorKey: 'transformer',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Transformer" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <FormField
              name={`mappings.${row.index}.transformer`}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <SelectTrigger className="w-[200px]">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {transformers?.map((t) => (
                          <SelectItem
                            className="cursor-pointer"
                            key={t.value}
                            value={t.value as unknown as string}
                          >
                            {t.title}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormControl>

                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        );
      },
    },
    {
      accessorKey: 'exclude',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Exclude" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <FormField
              name={`mappings.${row.index}.exclude`}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <Select
                      onValueChange={field.onChange}
                      value={
                        field.value == true || field.value == 'true'
                          ? 'true'
                          : 'false'
                      }
                    >
                      <SelectTrigger className="w-[125px]">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem
                          className="cursor-pointer"
                          key="exclude"
                          value="true"
                        >
                          Exclude
                        </SelectItem>

                        <SelectItem
                          className="cursor-pointer"
                          key="include"
                          value="false"
                        >
                          Include
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        );
      },
    },
  ];
}
