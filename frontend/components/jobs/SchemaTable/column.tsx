'use client';

import EditTransformerOptions, {
  handleTransformerMetadata,
} from '@/app/transformers/EditTransformerOptions';
import { Checkbox } from '@/components/ui/checkbox';

import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command';
import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

import { cn } from '@/libs/utils';
import { DatabaseColumn } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  CustomTransformer,
  Transformer,
  TransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { PlainMessage } from '@bufbuild/protobuf';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { useState } from 'react';
import { DataTableColumnHeader } from './data-table-column-header';

interface GetColumnsProps {
  systemTransformers: Transformer[];
  customTransformers: CustomTransformer[];
}

export function getColumns(
  props: GetColumnsProps
): ColumnDef<PlainMessage<DatabaseColumn>>[] {
  const { systemTransformers, customTransformers } = props;

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
          className="translate-y-[2px] "
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
      cell: ({ row }) => <div>{row.getValue('schema')}</div>,
      enableSorting: true,
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'table',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Table" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="truncate font-medium">
              {row.getValue('table')}
            </span>
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'column',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Column" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="truncate font-medium">
              {row.getValue('column')}
            </span>
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'dataType',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Data Type" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2 ">
            <span className="truncate font-medium">
              {row.getValue('dataType')}
            </span>
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
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
              name={`mappings.${row.index}.transformer.value`}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <div className="flex flex-row space-x-2">
                      <TransformerSelect
                        systemTransformers={systemTransformers}
                        customTransformers={customTransformers}
                        value={field.value}
                        onSelect={field.onChange}
                      />
                      <EditTransformerOptions
                        transformer={systemTransformers?.find(
                          (item) => item.value == field.value
                        )}
                        index={row.index}
                      />
                    </div>
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
        const rowValue = row.getValue(id) as Transformer;
        return value.includes(rowValue.value);
      },
    },
  ];
}

interface TransformersSelectProps {
  systemTransformers: Transformer[];
  customTransformers: CustomTransformer[];
  value: string;
  onSelect: (value: string) => void;
}

function TransformerSelect(props: TransformersSelectProps) {
  const { systemTransformers, customTransformers, value, onSelect } = props;
  const [open, setOpen] = useState(false);

  let merged: CustomTransformer[] = [...customTransformers];

  systemTransformers.map((st) => {
    const cf = {
      config: {
        case: st.config?.config.case,
        value: st.config?.config.value,
      },
    };

    const newCt = new CustomTransformer({
      name: handleTransformerMetadata(st.value).name,
      description: handleTransformerMetadata(st.value).description,
      type: handleTransformerMetadata(st.value).type,
      source: st.value,
      config: cf as TransformerConfig,
    });

    merged.push(newCt);
  });

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="justify-between  min-w-[141px] max-w-[141px]"
        >
          <div className="truncate overflow-hidden text-ellipsis whitespace-nowrap">
            {value}
          </div>
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className=" p-0">
        <Command>
          <CommandInput placeholder="Search transformers..." />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <CommandGroup className="overflow-auto h-[400px]">
            {merged.map((t, index) => (
              <CommandItem
                key={`${t.name}-${index}`}
                onSelect={() => {
                  onSelect(t.name);
                  setOpen(false);
                }}
                value={t.name}
              >
                <CheckIcon
                  className={cn(
                    'mr-2 h-4 w-4',
                    value == t.name ? 'opacity-100' : 'opacity-0'
                  )}
                />
                {t.name}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
