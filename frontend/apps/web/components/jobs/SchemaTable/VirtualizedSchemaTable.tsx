import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { TreeData, VirtualizedTree } from '@/components/VirtualizedTree';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
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
import { ExclamationTriangleIcon, UpdateIcon } from '@radix-ui/react-icons';
import { CSSProperties, ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List } from 'react-window';
import ColumnFilterSelect from './ColumnFilterSelect';
import TransformerSelect from './TransformerSelect';

export type Row = JobMappingFormValues & {
  isSelected: boolean;
  formIdx: number;
};

// columnId: list of values
type ColumnFilters = Record<string, string[]>;

interface VirtualizedSchemaTableProps {
  data: Row[];
  transformers: Transformer[];
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

  // this is sketch
  // useEffect(() => {
  //   setRows(data);
  // }, [data]);

  // const treeData = useMemo(
  //   () => getSchemaTreeData(data, columnFilters),
  //   [data, columnFilters]
  // );
  const treeData = getSchemaTreeData(data, columnFilters, transformers);

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
    setRows((prevItems) => {
      const newItems = [...prevItems];
      newItems[index] = {
        ...newItems[index],
        isSelected: !newItems[index].isSelected,
      };
      return newItems;
    });
  };

  const onSelectAll = (isSelected: boolean): void => {
    setBulkSelect(isSelected);
    setRows((prevItems) => {
      return [...prevItems].map((i) => {
        return {
          ...i,
          isSelected,
        };
      });
    });
  };

  const onTreeFilterSelect = (id: string, isSelected: boolean): void => {
    setColumnFilters((prevFilters) => {
      const [schema, table] = splitOnFirstOccurrence(id, '.');
      const newFilters = { ...prevFilters };
      if (isSelected) {
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
                  if (r.isSelected) {
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
        <div>
          <VirtualizedSchemaList
            filteredRows={rows}
            allRows={data}
            onSelect={onSelect}
            onSelectAll={onSelectAll}
            transformers={transformers}
            isAllSelected={bulkSelect}
            columnFilters={columnFilters}
            onFilterSelect={onFilterSelect}
          />
        </div>
      </div>
    </div>
  );
};

interface RowItemData {
  rows: Row[];
  onSelect: (index: number) => void;
  onSelectAll: (value: boolean) => void;
  transformers: Transformer[];
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
  const { rows, onSelect, transformers } = data;
  const row = rows[index];

  return (
    <div style={style} className="border-t dark:border-gray-700">
      <div className="grid grid-cols-5 gap-2 items-center p-2">
        <div className="flex flex-row truncate ">
          <Checkbox
            id="select"
            onClick={() => onSelect(index)}
            checked={row.isSelected}
            type="button"
            className="self-center mr-4"
          />
          <Cell value={row.schema} />
        </div>
        <Cell value={row.table} />
        <Cell value={row.column} />
        <Cell value={row.dataType} />
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
                            fv.config.case === 'userDefinedTransformerConfig' &&
                            isUserDefinedTransformer(t) &&
                            t.id === fv.config.value.id
                          ) {
                            return t;
                          }
                          return (
                            isSystemTransformer(t) && t.source === fv.source
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
// const createRowData = memoize(
//   (
//     rows: Row[],
//     onSelect: (index: number) => void,
//     onSelectAll: (value: boolean) => void,
//     transformers: Transformer[]
//   ) => ({
//     rows,
//     onSelect,
//     onSelectAll,
//     transformers,
//   })
// );

interface VirtualizedSchemaListProps {
  filteredRows: Row[];
  allRows: Row[];
  onSelect: (index: number) => void;
  onSelectAll: (isSelected: boolean) => void;
  isAllSelected: boolean;
  columnFilters: ColumnFilters;
  onFilterSelect: (columnId: string, newValues: string[]) => void;
  transformers: Transformer[];
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
}: VirtualizedSchemaListProps) {
  // Bundle additional data to list rows using the "rowData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  // const rowData = rows // createRowData(rows, onSelect, onSelectAll, transformers);
  // const uniqueFilters = useMemo(
  //   () => getUniqueFilters(allRows, columnFilters, transformers),
  //   [allRows, columnFilters]
  // );
  const uniqueFilters = getUniqueFilters(allRows, columnFilters, transformers);

  const sumOfRowHeights = 50 * filteredRows.length;
  const dynamicHeight = Math.min(sumOfRowHeights, 700);
  const itemData: RowItemData = {
    rows: filteredRows,
    onSelect,
    onSelectAll,
    transformers,
  };

  return (
    <div
      className={cn(`grid grid-col-1 border rounded-md dark:border-gray-700`)}
    >
      <div className={`grid grid-cols-5 gap-2 pl-2 pt-1 bg-muted `}>
        <div className="flex flex-row">
          <Checkbox
            id="select"
            onClick={() => {
              onSelectAll(!isAllSelected);
            }}
            checked={isAllSelected}
            type="button"
            className="self-center mr-4"
          />

          <span className="text-xs self-center">Schema</span>
          <ColumnFilterSelect
            columnId="schema"
            allColumnFilters={columnFilters}
            setColumnFilters={onFilterSelect}
            possibleFilters={uniqueFilters.schema}
          />
        </div>
        <div className="flex flex-row">
          <span className="text-xs self-center">Table</span>
          <ColumnFilterSelect
            columnId="table"
            allColumnFilters={columnFilters}
            setColumnFilters={onFilterSelect}
            possibleFilters={uniqueFilters.table}
          />
        </div>
        <div className="flex flex-row">
          <span className="text-xs self-center">Column</span>
          <ColumnFilterSelect
            columnId="column"
            allColumnFilters={columnFilters}
            setColumnFilters={onFilterSelect}
            possibleFilters={uniqueFilters.column}
          />
        </div>
        <div className="flex flex-row">
          <span className="text-xs self-center">Data Type</span>
          <ColumnFilterSelect
            columnId="dataType"
            allColumnFilters={columnFilters}
            setColumnFilters={onFilterSelect}
            possibleFilters={uniqueFilters.dataType}
          />
        </div>
        <div className="flex flex-row">
          <span className="text-xs self-center">Transformer</span>
          <ColumnFilterSelect
            columnId="transformer"
            allColumnFilters={columnFilters}
            setColumnFilters={onFilterSelect}
            possibleFilters={uniqueFilters.transformer}
          />
        </div>
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
