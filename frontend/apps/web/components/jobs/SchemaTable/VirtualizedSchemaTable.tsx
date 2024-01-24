import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { TreeData, VirtualizedTree } from '@/components/VirtualizedTree';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { FormControl, FormField, FormItem } from '@/components/ui/form';
import { cn } from '@/libs/utils';
import {
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import {
  JobMappingFormValues,
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { UserDefinedTransformerConfig } from '@neosync/sdk';
import {
  ChevronDownIcon,
  ExclamationTriangleIcon,
  UpdateIcon,
} from '@radix-ui/react-icons';
import memoizeOne from 'memoize-one';
import { CSSProperties, ReactElement, useMemo, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List } from 'react-window';
import ColumnFilterSelect from './ColumnFilterSelect';
import TransformerSelect from './TransformerSelect';

export type Row = JobMappingFormValues & {
  // isSelected: boolean;
  formIdx: number;
};

// columnId: list of values
type ColumnFilters = Record<string, string[]>;

interface VirtualizedSchemaTableProps {
  data: Row[];
  transformers: Transformer[];
}

interface Column {
  name: string;
  id: string;
}

function buildRowKey(row: Row): string {
  return `${row.schema}-${row.table}-${row.column}`;
}

export const VirtualizedSchemaTable = function VirtualizedSchemaTable({
  data,
  transformers,
}: VirtualizedSchemaTableProps) {
  const [rows, setRows] = useState<Row[]>(data);
  const [bulkTransformer, setBulkTransformer] =
    useState<JobMappingTransformerForm>({
      source: '',
      config: { case: '', value: {} },
    });
  const [bulkSelect, setBulkSelect] = useState(false);
  const [columnFilters, setColumnFilters] = useState<ColumnFilters>({});
  const form = useFormContext<SingleTableSchemaFormValues | SchemaFormValues>();
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const treeData = getSchemaTreeData(data, columnFilters, transformers);
  // const treeData = useMemo(
  //   () => getSchemaTreeData(data, columnFilters, transformers),
  //   [data, columnFilters]
  // );

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

  const onSelect = (index: number): void => {
    setSelected((prevSet) => {
      const newSet = new Set(prevSet);
      const row = rows[index];
      if (newSet.has(buildRowKey(row))) {
        newSet.delete(buildRowKey(row));
      } else {
        newSet.add(buildRowKey(row));
      }
      return newSet;
    });
  };

  const onSelectAll = (isSelected: boolean): void => {
    setBulkSelect(isSelected);
    setSelected(new Set());
  };

  const onTreeFilterSelect = (id: string, isSelected: boolean): void => {
    setColumnFilters((prevFilters) => {
      const [schema, table] = splitOnFirstOccurrence(id, '.');
      const newFilters = { ...prevFilters };
      if (isSelected || bulkSelect) {
        newFilters['schema'] = newFilters['schema']
          ? [...newFilters['schema'], schema]
          : [schema];
        if (table) {
          newFilters['table'] = newFilters['table']
            ? [...newFilters['table'], table]
            : [table];
        }
      } else {
        newFilters['schema'] = newFilters['schema'].filter((s) => s != schema);
        if (table) {
          newFilters['table'] = newFilters['table'].filter((t) => t != table);
        }
      }
      const filteredRows = data.filter((r) =>
        shouldFilterRow(r, newFilters, transformers)
      );
      setRows(filteredRows);
      return newFilters;
    });
  };

  // construct the column order
  const columnList: Column[] = [
    { name: 'Schema', id: 'schema' },
    { name: 'Table', id: 'table' },
    { name: 'Column', id: 'column' },
    { name: 'Data Type', id: 'dataType' },
    { name: 'Transformer', id: 'transformer' },
  ];

  const [visibleColumns, setVisibleColumns] = useState<Column[]>(columnList);

  return (
    <div className="flex flex-row w-full">
      <div className="basis-1/6  pt-[45px] ">
        <VirtualizedTree data={treeData} onNodeSelect={onTreeFilterSelect} />
      </div>
      <div className="space-y-2 pl-2 basis-5/6 ">
        <div className="flex items-center justify-between">
          <div className="w-[250px]">
            <TransformerSelect
              transformers={transformers}
              value={bulkTransformer}
              onSelect={(value) => {
                rows.forEach((r) => {
                  if (bulkSelect || selected.has(buildRowKey(r))) {
                    form.setValue(`mappings.${r.formIdx}.transformer`, value, {
                      shouldDirty: true,
                    });
                  }
                });
                onSelectAll(false);
                setBulkTransformer({
                  source: '',
                  config: { case: '', value: {} },
                });
              }}
              placeholder="Bulk update Transformers..."
            />
          </div>
          <div className="flex flex-row items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="ml-auto">
                  Columns <ChevronDownIcon className="ml-2 h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {columnList.map((column) => {
                  return (
                    <DropdownMenuCheckboxItem
                      key={column.id}
                      className="capitalize"
                      checked={visibleColumns.some(
                        (item) => column.id == item.id
                      )}
                      onCheckedChange={() => {
                        // checks if the column is already visible, if it is, then it filters it from the visible columns
                        if (
                          visibleColumns.some((item) => column.id == item.id)
                        ) {
                          const updatedColumns = visibleColumns.filter(
                            (item) => item.id != column.id
                          );
                          setVisibleColumns(updatedColumns);
                        } else {
                          // adds the column back into the visible columns in the same order as the original columns list
                          const newVisibleColumns = columnList.filter(
                            (col) =>
                              visibleColumns.some(
                                (item) => item.id === col.id
                              ) || col.id === column.id
                          );
                          setVisibleColumns(newVisibleColumns);
                        }
                      }}
                    >
                      {column.id}
                    </DropdownMenuCheckboxItem>
                  );
                })}
              </DropdownMenuContent>
            </DropdownMenu>
            <Button
              variant="outline"
              type="button"
              onClick={() => {
                setColumnFilters({});
                setRows(data);
              }}
            >
              Clear filters
              <UpdateIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            </Button>
          </div>
        </div>
        <div>
          <VirtualizedSchemaList
            filteredRows={rows}
            allRows={data}
            onSelect={onSelect}
            onSelectAll={onSelectAll}
            transformers={transformers}
            isAllSelected={bulkSelect}
            selected={selected}
            columnFilters={columnFilters}
            onFilterSelect={onFilterSelect}
            visibleColumns={visibleColumns}
          />
        </div>
      </div>
    </div>
  );
};

interface RowItemData {
  rows: Row[];
  selected: Set<string>;
  isAllSelected: boolean;
  onSelect: (index: number) => void;
  onSelectAll: (value: boolean) => void;
  transformers: Transformer[];
  visibleColumns: Column[];
}

interface TableRowProps {
  index: number;
  style: CSSProperties;
  data: RowItemData;
}

// If list items are expensive to render,
// Consider using PureComponent to avoid unnecessary re-renders.
// https://reactjs.org/docs/react-api.html#reactpurecomponent
const TableRow = function Row({ data, index, style }: TableRowProps) {
  // Data passed to List as "itemData" is available as props.data
  const {
    rows,
    onSelect,
    transformers,
    selected,
    isAllSelected,
    visibleColumns,
  } = data;
  const row = rows[index];

  return (
    <div style={style} className="border-t dark:border-gray-700">
      <div className="grid grid-cols-5 gap-2 items-center p-2">
        {visibleColumns.map((column) => {
          switch (column.id) {
            case 'schema':
              return (
                <div className="flex flex-row truncate " key={column.id}>
                  <Checkbox
                    id="select"
                    onClick={() => onSelect(index)}
                    checked={isAllSelected || selected.has(buildRowKey(row))}
                    type="button"
                    className="self-center mr-4"
                  />
                  <Cell value={row.schema} />
                </div>
              );
            case 'table':
              return <Cell key={column.id} value={row.table} />;
            case 'column':
              return <Cell key={column.id} value={row.column} />;
            case 'dataType':
              return <Cell key={column.id} value={row.dataType} />;
            case 'transformer':
              return (
                <div>
                  <FormField<SchemaFormValues | SingleTableSchemaFormValues>
                    name={`mappings.${row.formIdx}.transformer`}
                    render={({ field, fieldState, formState }) => {
                      const fv = field.value as JobMappingTransformerForm;
                      return (
                        <FormItem>
                          <FormControl>
                            <div className="flex flex-row space-x-2">
                              {formState.errors.mappings && (
                                <div className="place-self-center">
                                  {fieldState.error ? (
                                    <div>
                                      <ExclamationTriangleIcon className="h-4 w-4 text-destructive" />
                                    </div>
                                  ) : (
                                    <div className="w-4"></div>
                                  )}
                                </div>
                              )}

                              <div>
                                <TransformerSelect
                                  transformers={transformers}
                                  value={fv}
                                  onSelect={field.onChange}
                                  placeholder="Select Transformer..."
                                />
                              </div>
                              <EditTransformerOptions
                                transformer={transformers.find((t) => {
                                  if (
                                    fv.source === 'custom' &&
                                    fv.config.case ===
                                      'userDefinedTransformerConfig' &&
                                    isUserDefinedTransformer(t) &&
                                    t.id === fv.config.value.id
                                  ) {
                                    return t;
                                  }
                                  return (
                                    isSystemTransformer(t) &&
                                    t.source === fv.source
                                  );
                                })}
                                index={index}
                              />
                            </div>
                          </FormControl>
                        </FormItem>
                      );
                    }}
                  />
                </div>
              );
            default:
              return null;
          }
        })}
      </div>
    </div>
  );
};
TableRow.displayName = 'row';

interface CellProps {
  value: string;
}

function Cell(props: CellProps): ReactElement {
  const { value } = props;
  return <span className="truncate font-medium text-sm">{value}</span>;
}

// This helper function memoizes incoming props,
// To avoid causing unnecessary re-renders pure Row components.
// This is only needed since we are passing multiple props with a wrapper object.
// If we were only passing a single, stable value (e.g. items),
// We could just pass the value directly.
const createRowData = memoizeOne(
  (
    rows: Row[],
    onSelect: (index: number) => void,
    onSelectAll: (value: boolean) => void,
    transformers: Transformer[],
    selected: Set<string>,
    isAllSelected: boolean,
    visibleColumns: Column[]
  ) => ({
    rows,
    onSelect,
    onSelectAll,
    transformers,
    selected,
    isAllSelected,
    visibleColumns,
  })
);

interface VirtualizedSchemaListProps {
  filteredRows: Row[];
  allRows: Row[];
  onSelect: (index: number) => void;
  onSelectAll: (isSelected: boolean) => void;
  isAllSelected: boolean;
  selected: Set<string>;
  columnFilters: ColumnFilters;
  onFilterSelect: (columnId: string, newValues: string[]) => void;
  transformers: Transformer[];
  visibleColumns: Column[];
}
// In this example, "items" is an Array of objects to render,
// and "onSelect" is a function that updates an item's state.
function VirtualizedSchemaList({
  filteredRows,
  allRows,
  onSelect,
  onSelectAll,
  transformers,
  isAllSelected,
  columnFilters,
  onFilterSelect,
  selected,
  visibleColumns,
}: VirtualizedSchemaListProps) {
  // Bundle additional data to list rows using the "rowData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  const itemData: RowItemData = createRowData(
    filteredRows,
    onSelect,
    onSelectAll,
    transformers,
    selected,
    isAllSelected,
    visibleColumns
  );

  const uniqueFilters = useMemo(
    () => getUniqueFilters(allRows, columnFilters, transformers),
    [allRows, columnFilters]
  );

  const sumOfRowHeights = 50 * filteredRows.length;
  const dynamicHeight = Math.min(sumOfRowHeights, 700);

  return (
    <div
      className={cn(`grid grid-col-1 border rounded-md dark:border-gray-700`)}
    >
      <div className={`grid grid-cols-5 gap-2 pl-2 pt-1 bg-muted `}>
        {visibleColumns.map((col) => (
          <div className="flex flex-row" key={col.id}>
            {col.id == 'schema' && (
              <Checkbox
                id="select"
                onClick={() => {
                  onSelectAll(!isAllSelected);
                }}
                checked={isAllSelected}
                type="button"
                className="self-center mr-4"
              />
            )}
            <span className="text-xs self-center">{col.name}</span>
            <ColumnFilterSelect
              columnId={col.id}
              allColumnFilters={columnFilters}
              setColumnFilters={onFilterSelect}
              possibleFilters={uniqueFilters[col.id]}
            />
          </div>
        ))}
        <div className="col-span-5"></div>
      </div>

      <div style={{ height: dynamicHeight }}>
        <AutoSizer defaultHeight={700}>
          {({ height, width }) => (
            <List
              height={height}
              itemCount={filteredRows.length}
              itemData={itemData}
              itemSize={50}
              width={width}
              itemKey={(index: number) => {
                const r = filteredRows[index];
                return buildRowKey(r);
              }}
            >
              {TableRow}
            </List>
          )}
        </AutoSizer>
      </div>
    </div>
  );
}

function shouldFilterRow(
  row: Row,
  columnFilters: ColumnFilters,
  transformers: Transformer[],
  columnIdToSkip?: keyof Row
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
        const rowVal = row[key as keyof Row] as JobMappingTransformerForm;
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
        const value = row[key as keyof Row] as string;
        if (!filters.includes(value)) {
          return false;
        }
      }
    }
  }
  return true;
}

function isTableSelected(
  table: string,
  schema: string,
  tableFilters: Set<string>,
  schemaFilters: Set<string>
): boolean {
  return (
    (schemaFilters.size === 0 || schemaFilters.has(schema)) &&
    (tableFilters.size === 0 || tableFilters.has(table))
  );
}

function getSchemaTreeData(
  data: Row[],
  columnFilters: ColumnFilters,
  transformers: Transformer[]
): TreeData[] {
  const schemaMap: Record<string, Record<string, string>> = {};
  data.forEach((row) => {
    // don't build the map if there are no rows that are filterable by the schema or table
    if (
      !shouldFilterRow(row, columnFilters, transformers, 'schema') &&
      !shouldFilterRow(row, columnFilters, transformers, 'table')
    ) {
      return;
    }
    if (!schemaMap[row.schema]) {
      schemaMap[row.schema] = { [row.table]: row.table };
    } else {
      schemaMap[row.schema][row.table] = row.table;
    }
  });

  const schemaFilters = new Set(columnFilters['schema'] || []);
  const tableFilters = new Set(columnFilters['table'] || []);

  const falseOverride = schemaFilters.size === 0 && tableFilters.size === 0;

  return Object.keys(schemaMap).map((schema): TreeData => {
    const isSchemaSelected = schemaFilters.has(schema);
    const children = Object.keys(schemaMap[schema]).map((table): TreeData => {
      return {
        id: `${schema}.${table}`,
        name: table,
        isSelected: falseOverride
          ? false
          : isTableSelected(table, schema, tableFilters, schemaFilters),
      };
    });
    const isSomeTablesSelected = children.some((t) => t.isSelected);

    return {
      id: schema,
      name: schema,
      isSelected: falseOverride
        ? false
        : isSchemaSelected || isSomeTablesSelected,
      children,
    };
  });
}

function getUniqueFilters(
  allRows: Row[],
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
  row: Row,
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

function splitOnFirstOccurrence(
  str: string,
  character: string
): [string, string] {
  const index = str.indexOf(character);

  if (index === -1) {
    // Character not found, return the original string and an empty string
    return [str, ''];
  }

  // Split the string at the found index
  const firstPart = str.substring(0, index);
  const secondPart = str.substring(index + 1); // +1 to exclude the character itself

  return [firstPart, secondPart];
}
