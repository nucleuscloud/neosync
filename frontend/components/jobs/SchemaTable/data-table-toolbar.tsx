'use client';

import { Table } from '@tanstack/react-table';

import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/libs/utils';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { CaretSortIcon, CheckIcon, UpdateIcon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  transformers?: Transformer[];
  onClearFilters: () => void;
}

export function DataTableToolbar<TData>({
  table,
  transformers,
  onClearFilters,
}: DataTableToolbarProps<TData>) {
  const form = useFormContext();
  const [transformer, setTransformer] = useState<string>('');
  const [exclude, setExclude] = useState<string>('');

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2">
        <BulkTansformerSelect
          transformers={transformers || []}
          value={transformer}
          onSelect={(value) => {
            const rows = table.getSelectedRowModel();
            rows.rows.forEach((r) => {
              form.setValue(`mappings.${r.index}.transformer`, value, {
                shouldDirty: true,
              });
            });

            table.resetRowSelection();
            setTransformer('');
          }}
        />
        <Select
          value={exclude}
          onValueChange={(value) => {
            const rows = table.getSelectedRowModel();

            rows.rows.forEach((r) => {
              form.setValue(`mappings.${r.index}.exclude`, value, {
                shouldDirty: true,
              });
            });

            table.resetRowSelection();
            setExclude('');
          }}
        >
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="bulk update exclude..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem className="cursor-pointer" key="exclude" value="true">
              Exclude
            </SelectItem>

            <SelectItem className="cursor-pointer" key="include" value="false">
              Include
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <Button
        variant="outline"
        type="button"
        onClick={() => {
          table.setColumnFilters([]);
          onClearFilters();
        }}
      >
        Clear filters
        <UpdateIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </Button>
    </div>
  );
}

interface BulkTransformersSelectProps {
  transformers: Transformer[];
  value: string;
  onSelect: (value: string) => void;
}

function BulkTansformerSelect(props: BulkTransformersSelectProps) {
  const { transformers, value, onSelect } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-[250px] justify-between"
        >
          {value
            ? transformers.find((t) => t.value === value)?.title
            : 'Bulk update transformers...'}
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[250px] p-0">
        <Command>
          <CommandInput placeholder="Search transformers..." />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <CommandGroup>
            {transformers.map((t, index) => (
              <CommandItem
                key={`${t.value}-${index}`}
                onSelect={(currentValue) => {
                  onSelect(currentValue);
                  setOpen(false);
                }}
                value={t.value}
              >
                <CheckIcon
                  className={cn(
                    'mr-2 h-4 w-4',
                    value == t.value ? 'opacity-100' : 'opacity-0'
                  )}
                />
                {t.title}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
