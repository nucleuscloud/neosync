import { SingleTableSchemaFormValues } from '@/app/[account]/new/job/schema';
import EditTransformerOptions from '@/app/[account]/transformers/EditTransformerOptions';
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
  SchemaFormValues,
  TransformerFormValues,
} from '@/yup-validations/jobs';
import {
  JobMappingTransformer,
  UserDefinedTransformerConfig,
} from '@neosync/sdk';
import { ExclamationTriangleIcon, UpdateIcon } from '@radix-ui/react-icons';
import memoize from 'memoize-one';
import {
  CSSProperties,
  memo,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { useFormContext } from 'react-hook-form';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List, areEqual } from 'react-window';
import { VirtualizedTree } from '../../VirtualizedTree';
import ColumnFilterSelect from './ColumnFilterSelect';
import TransformerSelect from './TransformerSelect';

interface Row {
  table: string;
  transformer: TransformerFormValues;
  schema: string;
  column: string;
  dataType: string;
  isSelected: boolean;
}

type ColumnFilters = Record<string, string[]>;

interface VirtualizedSchemaTableProps {
  data: Row[];
  transformers: Transformer[];
}

export const VirtualizedSchemaTable = memo(function VirtualizedSchemaTable({
  data,
  transformers,
}: VirtualizedSchemaTableProps) {
  const [rows, setRows] = useState(data);
  const [bulkTransformer, setBulkTransformer] = useState<JobMappingTransformer>(
    new JobMappingTransformer({})
  );
  const [bulkSelect, setBulkSelect] = useState(false);
  const [columnFilters, setColumnFilters] = useState<ColumnFilters>({});
  const form = useFormContext<SingleTableSchemaFormValues | SchemaFormValues>();

  useEffect(() => {
    setRows(data);
  }, [data]);

  const treeData = useMemo(
    () => getSchemaTreeData(data, columnFilters),
    [data, columnFilters]
  );

  const onFilterSelect = useCallback(
    (columnId: string, colFilters: string[]) => {
      setColumnFilters((prevFilters) => {
        const newFilters = { ...prevFilters, [columnId]: colFilters };
        if (colFilters.length == 0) {
          delete newFilters[columnId as keyof ColumnFilters];
        }
        const filteredRows = data.filter((r) =>
          shouldFilterRow(r, newFilters, transformers)
        );
        setRows(filteredRows);
        return newFilters;
      });
    },
    []
  );

  const onSelect = useCallback((index: number) => {
    setRows((prevItems) => {
      const newItems = [...prevItems];
      newItems[index] = {
        ...newItems[index],
        isSelected: !newItems[index].isSelected,
      };
      return newItems;
    });
  }, []);

  const onSelectAll = useCallback((isSelected: boolean) => {
    setRows((prevItems) => {
      const newItems = [...prevItems];
      return newItems.map((i) => {
        return {
          ...i,
          isSelected,
        };
      });
    });
  }, []);

  const onTreeFilterSelect = useCallback((id: string, isSelected: boolean) => {
    setColumnFilters((prevFilters) => {
      const [schema, table] = id.split('.');
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
  }, []);

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
                rows.forEach((r, index) => {
                  if (r.isSelected) {
                    form.setValue(
                      `mappings.${index}.transformer`,
                      {
                        source: value.source,
                        config: {
                          config: {
                            value: value.config?.config.value!,
                            case: value.config?.config.case,
                          },
                        },
                      },
                      {
                        shouldDirty: true,
                      }
                    );
                  }
                });
                onSelectAll(false);
                setBulkSelect(false);
                setBulkTransformer(value);
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
            rows={rows}
            allRows={data}
            onSelect={onSelect}
            onSelectAll={onSelectAll}
            transformers={transformers}
            bulkSelect={bulkSelect}
            setBulkSelect={setBulkSelect}
            columnFilters={columnFilters}
            onFilterSelect={onFilterSelect}
          />
        </div>
      </div>
    </div>
  );
});

interface RowProps {
  index: number;
  style: CSSProperties;
  data: {
    rows: Row[];
    onSelect: (index: number) => void;
    onSelectAll: (value: boolean) => void;
    transformers: Transformer[];
  };
}

// If list items are expensive to render,
// Consider using PureComponent to avoid unnecessary re-renders.
// https://reactjs.org/docs/react-api.html#reactpurecomponent
const Row = memo(function Row({ data, index, style }: RowProps) {
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
            name={`mappings.${index}.transformer`}
            render={({ field, fieldState, formState }) => {
              // todo: we should really convert between the real field.value and the job mapping transformer
              const fv = field.value as unknown as JobMappingTransformer;

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

                      <div className="">
                        <TransformerSelect
                          transformers={transformers}
                          value={fv}
                          onSelect={field.onChange}
                          placeholder="Select transformer..."
                        />
                      </div>
                      <EditTransformerOptions
                        transformer={transformers.find((t) => {
                          if (
                            fv.source === 'custom' &&
                            fv.config?.config.case ===
                              'userDefinedTransformerConfig' &&
                            isUserDefinedTransformer(t) &&
                            t.id === fv.config?.config.value.id
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
}, areEqual);
Row.displayName = 'row';

interface CellProps {
  value: string;
}

function Cell(props: CellProps) {
  const { value } = props;
  return <span className="truncate font-medium text-sm">{value}</span>;
}

// This helper function memoizes incoming props,
// To avoid causing unnecessary re-renders pure Row components.
// This is only needed since we are passing multiple props with a wrapper object.
// If we were only passing a single, stable value (e.g. items),
// We could just pass the value directly.
const createRowData = memoize(
  (
    rows: Row[],
    onSelect: (index: number) => void,
    onSelectAll: (value: boolean) => void,
    transformers: Transformer[]
  ) => ({
    rows,
    onSelect,
    onSelectAll,
    transformers,
  })
);

interface VirtualizedSchemaListProps {
  rows: Row[];
  allRows: Row[];
  onSelect: (index: number) => void;
  onSelectAll: (value: boolean) => void;
  bulkSelect: boolean;
  setBulkSelect: (value: boolean) => void;
  columnFilters: ColumnFilters;
  onFilterSelect: (columnId: string, newValues: string[]) => void;
  transformers: Transformer[];
}
// In this example, "items" is an Array of objects to render,
// and "onSelect" is a function that updates an item's state.
function VirtualizedSchemaList({
  rows,
  allRows,
  onSelect,
  onSelectAll,
  transformers,
  bulkSelect,
  setBulkSelect,
  columnFilters,
  onFilterSelect,
}: VirtualizedSchemaListProps) {
  // Bundle additional data to list rows using the "rowData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  const rowData = createRowData(rows, onSelect, onSelectAll, transformers);
  const uniqueFilters = useMemo(
    () => getUniqueFilters(allRows, columnFilters, transformers),
    [allRows, columnFilters]
  );

  const sumOfRowHeights = 50 * rows.length;
  const dynamicHeight = Math.min(sumOfRowHeights, 700);

  return (
    <div
      className={cn(`grid grid-col-1 border rounded-md dark:border-gray-700`)}
    >
      <div className={`grid grid-cols-5 gap-2 pl-2 pt-1 bg-muted `}>
        <div className="flex flex-row">
          <Checkbox
            id="select"
            onClick={() => {
              onSelectAll(!bulkSelect);
              setBulkSelect(!bulkSelect);
            }}
            checked={bulkSelect}
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
              itemCount={rows.length}
              itemData={rowData}
              itemSize={50}
              width={width}
              itemKey={(index: number) => {
                const r = rows[index];
                return `${r.schema}-${r.table}-${r.column}-${index}`;
              }}
            >
              {Row}
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
  columnId?: keyof Row
): boolean {
  for (const key of Object.keys(columnFilters)) {
    if (columnId && key == columnId) {
      continue;
    }
    const filters = columnFilters[key as keyof ColumnFilters];
    if (filters.length == 0) {
      continue;
    }
    switch (key) {
      case 'transformer': {
        const rowVal = row[key as keyof Row] as JobMappingTransformer;
        if (rowVal.source === 'custom') {
          const udfId = (
            rowVal.config?.config.value as UserDefinedTransformerConfig
          ).id;
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

function getSchemaTreeData(data: Row[], columnFilters: ColumnFilters) {
  const schemaMap: Record<string, Record<string, string>> = {};
  data.forEach((row) => {
    if (!schemaMap[row.schema]) {
      schemaMap[row.schema] = { [row.table]: row.table };
    } else {
      schemaMap[row.schema][row.table] = row.table;
    }
  });

  const schemaFilters = new Set(columnFilters['schema'] || []);
  const tableFilters = new Set(columnFilters['table'] || []);

  var falseOverride = false;
  if (schemaFilters.size == 0 && tableFilters.size == 0) {
    falseOverride = true;
  }

  return Object.keys(schemaMap).map((schema) => {
    const isSchemaSelected = schemaFilters.has(schema);
    const children = Object.keys(schemaMap[schema]).map((table) => {
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

function getUniqueFiltersByColumn(
  rows: Row[],
  columnFilters: ColumnFilters,
  columnId: keyof Row,
  transformers: Transformer[]
): string[] {
  const uniqueColFilters: Record<string, string> = {};
  const filteredRows = rows.filter((r) =>
    shouldFilterRow(r, columnFilters, transformers, columnId)
  );
  filteredRows.forEach((r) => {
    switch (columnId) {
      case 'transformer': {
        const rowVal = r[columnId];
        if (rowVal.source === 'custom') {
          const udfId = (
            rowVal.config.config.value as UserDefinedTransformerConfig
          ).id;
          const value =
            transformers.find(
              (t) => isUserDefinedTransformer(t) && t.id === udfId
            )?.name ?? 'unknown transformer';
          uniqueColFilters[value] = value;
        } else {
          const value =
            transformers.find(
              (t) => isSystemTransformer(t) && t.source === rowVal.source
            )?.name ?? 'unknown transformer';
          uniqueColFilters[value] = value;
        }
        break;
      }
      case 'isSelected': {
        uniqueColFilters[r[columnId] ? 'true' : 'false'];
        break;
      }
      default: {
        const value = r[columnId] as string;
        uniqueColFilters[value] = value;
      }
    }
  });
  return Object.keys(uniqueColFilters).sort();
}

function getUniqueFilters(
  rows: Row[],
  columnFilters: ColumnFilters,
  transformers: Transformer[]
): Record<string, string[]> {
  const filterSet: Record<string, string[]> = {
    schema: getUniqueFiltersByColumn(rows, columnFilters, 'schema', []),
    table: getUniqueFiltersByColumn(rows, columnFilters, 'table', []),
    column: getUniqueFiltersByColumn(rows, columnFilters, 'column', []),
    dataType: getUniqueFiltersByColumn(rows, columnFilters, 'dataType', []),
    transformer: getUniqueFiltersByColumn(
      rows,
      columnFilters,
      'transformer',
      transformers
    ),
  };
  return filterSet;
}
