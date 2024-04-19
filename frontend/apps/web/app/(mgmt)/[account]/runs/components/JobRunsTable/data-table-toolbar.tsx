'use client';

import { Cross2Icon, ReloadIcon } from '@radix-ui/react-icons';
import { Table } from '@tanstack/react-table';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

import Spinner from '@/components/Spinner';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { DataTableViewOptions } from './data-table-view-options';

interface DataTableToolbarProps<TData, TAutoRefreshInterval extends string> {
  table: Table<TData>;
  onRefreshClick(): void;
  refreshInterval: TAutoRefreshInterval;
  autoRefreshIntervalOptions: TAutoRefreshInterval[];
  onAutoRefreshIntervalChange(interval: TAutoRefreshInterval): void;
  isRefreshing: boolean;
}

export function DataTableToolbar<TData, TAutoRefreshInterval extends string>({
  table,
  onRefreshClick,
  refreshInterval,
  autoRefreshIntervalOptions,
  onAutoRefreshIntervalChange,
  isRefreshing,
}: DataTableToolbarProps<TData, TAutoRefreshInterval>) {
  const isFiltered = table.getState().columnFilters.length > 0;
  return (
    <div className="flex items-center justify-between space-x-4">
      <div className="flex flex-1 items-center space-x-2">
        <Input
          placeholder="Filter by Job Name..."
          value={(table.getColumn('jobName')?.getFilterValue() as string) ?? ''}
          onChange={(event) =>
            table.getColumn('jobName')?.setFilterValue(event.target.value)
          }
          className="h-8 w-[150px] lg:w-[350px]"
        />
        {isFiltered && (
          <Button
            variant="ghost"
            onClick={() => table.resetColumnFilters()}
            className="h-8 px-2 lg:px-3"
          >
            Reset
            <Cross2Icon className="ml-2 h-4 w-4" />
          </Button>
        )}
      </div>
      {isRefreshing && <Spinner />}
      <Select
        onValueChange={onAutoRefreshIntervalChange}
        defaultValue={refreshInterval}
      >
        <SelectTrigger className="w-[80px]">
          <SelectValue placeholder="Select auto-refresh interval" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>Auto-refresh interval</SelectLabel>
            {autoRefreshIntervalOptions.map((interval, index) => {
              return (
                <SelectItem key={`${interval}-${index}`} value={interval}>
                  {interval}
                </SelectItem>
              );
            })}
          </SelectGroup>
        </SelectContent>
      </Select>
      <Button variant="outline" size="icon" onClick={() => onRefreshClick()}>
        <ReloadIcon className="h-4 w-4" />
      </Button>
      <DataTableViewOptions table={table} />
    </div>
  );
}
