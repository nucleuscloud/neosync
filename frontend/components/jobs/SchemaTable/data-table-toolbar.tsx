'use client';

import { Table } from '@tanstack/react-table';

import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { UpdateIcon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  transformers?: Transformer[];
}

export function DataTableToolbar<TData>({
  table,
  transformers,
}: DataTableToolbarProps<TData>) {
  const form = useFormContext();
  const [transformer, setTransformer] = useState<string>('');
  const [exclude, setExclude] = useState<string>('');

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2">
        <Select
          onValueChange={(value) => {
            const rows = table.getSelectedRowModel();
            rows.rows.forEach((r) => {
              form.setValue(`mappings.${r.index}.transformer`, value, {
                shouldDirty: true,
              });
            });

            table.resetRowSelection();
            setTransformer('');
          }}
          value={transformer}
        >
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="bulk update transformer..." />
          </SelectTrigger>
          <SelectContent>
            {transformers?.map((t) => (
              <SelectItem
                className="cursor-pointer"
                key={t.value}
                value={t.value}
              >
                {t.title}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
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
        onClick={() => table.setColumnFilters([])}
      >
        Clear filters
        <UpdateIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </Button>
    </div>
  );
}
