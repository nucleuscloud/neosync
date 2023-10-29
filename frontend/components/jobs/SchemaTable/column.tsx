'use client';

import EditTransformerOptions from '@/app/transformers/EditTransformerOptions';
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

  const mergedTransformers = MergeSystemAndCustomTransformers(
    systemTransformers,
    customTransformers
  );

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
                        mergedTransformers={mergedTransformers}
                        value={field.value}
                        onSelect={field.onChange}
                      />
                      <EditTransformerOptions
                        transformer={mergedTransformers?.find(
                          (item) => item.name == field.value
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
  mergedTransformers: CustomTransformer[];
  value: string;
  onSelect: (value: string) => void;
}

function TransformerSelect(props: TransformersSelectProps) {
  const { mergedTransformers, value, onSelect } = props;
  const [open, setOpen] = useState(false);

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
            {mergedTransformers.map((t, index) => (
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

// merge system into custom and add in additional metadata fields for system transformers
// to fit into the custom transformers interface
export function MergeSystemAndCustomTransformers(
  system: Transformer[],
  custom: CustomTransformer[]
): CustomTransformer[] {
  let merged: CustomTransformer[] = [...custom];

  system.map((st) => {
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

  return merged;
}

interface TransformerMetadata {
  name: string;
  description: string;
  type: string;
}

export function handleTransformerMetadata(
  value: string | undefined
): TransformerMetadata {
  const tEntries: Record<string, TransformerMetadata>[] = [
    {
      email: {
        name: 'Email',
        description: 'Anonymizes or generates a new email.',
        type: 'string',
      },
    },
    {
      phone_number: {
        name: 'Phone Number',
        description:
          'Anonymizes or generates a new phone number. The default format is <XXX-XXX-XXXX>.',
        type: 'string',
      },
    },
    {
      int_phone_number: {
        name: 'Int64 Phone Number',
        description:
          'Anonymizes or generates a new phone number of type int64 with a default length of 10.',
        type: 'int64',
      },
    },
    {
      first_name: {
        name: 'First Name',
        description: 'Anonymizes or generates a new first name.',
        type: 'string',
      },
    },
    {
      last_name: {
        name: 'Last Name',
        description: 'Anonymizes or generates a new last name.',
        type: 'string',
      },
    },
    {
      full_name: {
        name: 'Full Name',
        description:
          'Anonymizes or generates a new full name consisting of a first and last name.',
        type: 'string',
      },
    },
    {
      uuid: {
        name: 'UUID',
        description: 'Generates a new UUIDv4 id.',
        type: 'uuid',
      },
    },
    {
      passthrough: {
        name: 'Passthrough',
        description:
          'Passes the input value through to the desination with no changes.',
        type: 'passthrough',
      },
    },
    {
      null: {
        name: 'Null',
        description: 'Inserts a <null> string instead of the source value.',
        type: 'null',
      },
    },
    {
      random_string: {
        name: 'Random String',
        description:
          'Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length or Preserve Length parameters are defined.',
        type: 'string',
      },
    },
    {
      random_bool: {
        name: 'Random Bool',
        description: 'Generates a boolean value at random.',
        type: 'bool',
      },
    },
    {
      random_int: {
        name: 'Random Integer',
        description:
          'Generates a random integer value with a default length of 4 unless the Integer Length or Preserve Length paramters are defined. .',
        type: 'int64',
      },
    },
    {
      random_float: {
        name: 'Random Float',
        description:
          'Generates a random float value with a default length of <XX.XXX>.',
        type: 'float',
      },
    },
    {
      gender: {
        name: 'Gender',
        description:
          'Randomly generates one of the following genders: female, male, undefined, nonbinary.',
        type: 'string',
      },
    },
    {
      utc_timestamp: {
        name: 'UTC Timestamp',
        description: 'Randomly generates a UTC timestamp.',
        type: 'time',
      },
    },
    {
      unix_timestamp: {
        name: 'Unix Timestamp',
        description: 'Randomly generates a Unix timestamp.',
        type: 'int64',
      },
    },
    {
      street_address: {
        name: 'Street Address',
        description:
          'Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor}. For example, 123 Main Street.',
        type: 'string',
      },
    },
    {
      city: {
        name: 'City',
        description:
          'Randomly selects a city from a list of predefined US cities.',
        type: 'string',
      },
    },
    {
      zipcode: {
        name: 'Zip Code',
        description:
          'Randomly selects a zip code from a list of predefined US cities.',
        type: 'string',
      },
    },
    {
      state: {
        name: 'State',
        description:
          'Randomly selects a US state and returns the two-character state code.',
        type: 'string',
      },
    },
    {
      full_address: {
        name: 'Full Address',
        description:
          'Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor} {city}, {state} {zipcode}. For example, 123 Main Street Boston, Massachusetts 02169. ',
        type: 'string',
      },
    },
  ];

  const def = {
    default: {
      name: 'Undefined',
      description: 'Undefined Transformer',
      type: 'undefined',
    },
  };

  if (!value) {
    return def.default;
  }
  const res = tEntries.find((item) => item[value]);

  return res ? res[value] : def.default;
}
