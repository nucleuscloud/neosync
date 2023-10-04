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
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
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
      enableColumnFilter: false,
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
      enableColumnFilter: true,
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
      enableColumnFilter: true,
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
      enableColumnFilter: true,
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
      enableColumnFilter: true,
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
                    <Select
                      onValueChange={field.onChange}
                      value={field.value == '' ? undefined : field.value}
                    >
                      <SelectTrigger className="w-[200px]">
                        <SelectValue placeholder="select a transformer..." />
                      </SelectTrigger>
                      <SelectContent>
                        {transformers?.map((t) => (
                          <SelectItem
                            className="cursor-pointer"
                            key={t.id}
                            value={t.name} //this is what gets sent to the backend, change to t.id to send the transformer id
                          >
                            {t.name}
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
      enableColumnFilter: true,
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
      enableColumnFilter: true,
      filterFn: (row, id, value) => {
        const filter = row.getValue(id) ? 'exclude' : 'include';
        return filter.includes(value);
      },
    },
  ];
}
