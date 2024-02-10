import { Column, Header, Table } from '@tanstack/react-table';

import { cn } from '@/libs/utils';
import {
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import {
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { UserDefinedTransformerConfig } from '@neosync/sdk';
import { useMemo, useState } from 'react';
import ColumnFilterSelect from '../SchemaTable/ColumnFilterSelect';
import { JobMapRow } from './makeData';

interface DataTableColumnHeaderProps<TData, TValue>
  extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>;
  title: string;
  table: Table<TData>;
  header: Header<JobMapRow, unknown>;
  data: DataRow[];
  transformers: Transformer[];
}

// columnId: list of values
type ColumnFilters = Record<string, string[]>;

export type DataRow = JobMappingFormValues & {
  // isSelected: boolean;
  formIdx: number;
};

export function SchemaTableColumnHeader<TData, TValue>({
  column,
  title,
  className,
  data,
  transformers,
}: DataTableColumnHeaderProps<TData, TValue>) {
  const [columnFilters, setColumnFilters] = useState<ColumnFilters>({});
  const [_, setRows] = useState<DataRow[]>(data);

  const onFilterSelect = (columnId: string, colFilters: string[]): void => {
    setColumnFilters((prevFilters) => {
      const newFilters = { ...prevFilters, [columnId]: colFilters };
      if (colFilters.length === 0) {
        delete newFilters[columnId as keyof ColumnFilters];
      }
      const filteredRows = data.filter((r) =>
        shouldFilterRow(r, newFilters, transformers)
      );
      setRows(filteredRows);
      return newFilters;
    });
  };

  const uniqueFilters = useMemo(
    () => getUniqueFilters(data, columnFilters, transformers),
    [data, columnFilters]
  );

  if (!column.getCanSort()) {
    return <div className={cn(className, 'text-xs')}>{title}</div>;
  }

  return (
    <>
      {column.id}
      <ColumnFilterSelect
        columnId={column.id}
        allColumnFilters={columnFilters}
        setColumnFilters={onFilterSelect}
        possibleFilters={uniqueFilters.column} // shoujld be uniqueFilters.{columnName}
      />
    </>
  );
}

function shouldFilterRow(
  row: DataRow,
  columnFilters: ColumnFilters,
  transformers: Transformer[],
  columnIdToSkip?: keyof DataRow
): boolean {
  for (const key of Object.keys(columnFilters)) {
    if (columnIdToSkip && key === columnIdToSkip) {
      continue;
    }
    const filters = columnFilters[key as keyof ColumnFilters];
    if (filters.length === 0) {
      continue;
    }
    switch (key) {
      case 'transformer': {
        const rowVal = row[key as keyof DataRow] as JobMappingTransformerForm;
        if (rowVal.source === 'custom') {
          const udfId = (rowVal.config.value as UserDefinedTransformerConfig)
            .id;
          const value =
            transformers.find(
              (t) => isUserDefinedTransformer(t) && t.id === udfId
            )?.name ?? 'unknown transformer';
          if (!filters.includes(value)) {
            return false;
          }
        } else {
          const value =
            transformers.find(
              (t) => isSystemTransformer(t) && t.source === rowVal.source
            )?.name ?? 'unknown transformer';
          if (!filters.includes(value)) {
            return false;
          }
        }
        break;
      }
      default: {
        const value = row[key as keyof DataRow] as string;
        if (!filters.includes(value)) {
          return false;
        }
      }
    }
  }
  return true;
}

function getUniqueFilters(
  allRows: DataRow[],
  columnFilters: ColumnFilters,
  transformers: Transformer[]
): Record<string, string[]> {
  const filterSet = {
    schema: new Set<string>(),
    table: new Set<string>(),
    column: new Set<string>(),
    dataType: new Set<string>(),
    transformer: new Set<string>(),
  };
  allRows.forEach((row) => {
    if (shouldFilterRow(row, columnFilters, transformers, 'schema')) {
      filterSet.schema.add(row.schema);
    }
    if (shouldFilterRow(row, columnFilters, transformers, 'table')) {
      filterSet.table.add(row.table);
    }
    if (shouldFilterRow(row, columnFilters, transformers, 'column')) {
      filterSet.column.add(row.column);
    }
    if (shouldFilterRow(row, columnFilters, transformers, 'dataType')) {
      filterSet.dataType.add(row.dataType);
    }
    if (shouldFilterRow(row, columnFilters, transformers, 'transformer')) {
      filterSet.transformer.add(getTransformerFilterValue(row, transformers));
    }
  });
  const uniqueColFilters: Record<string, string[]> = {};
  Object.entries(filterSet).forEach(([key, val]) => {
    uniqueColFilters[key] = Array.from(val).sort();
  });
  return uniqueColFilters;
}

function getTransformerFilterValue(
  row: DataRow,
  transformers: Transformer[]
): string {
  if (row.transformer.source === 'custom') {
    const udfId = (row.transformer.config.value as UserDefinedTransformerConfig)
      .id;
    return (
      transformers.find((t) => isUserDefinedTransformer(t) && t.id === udfId)
        ?.name ?? 'unknown transformer'
    );
  } else {
    return (
      transformers.find(
        (t) => isSystemTransformer(t) && t.source === row.transformer.source
      )?.name ?? 'unknown transformer'
    );
  }
}
