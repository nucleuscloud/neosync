import EditTransformerOptions from '@/app/transformers/EditTransformerOptions';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { FormControl, FormField, FormItem } from '@/components/ui/form';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { UpdateIcon } from '@radix-ui/react-icons';
import memoize from 'memoize-one';
import { memo, useCallback, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { FixedSizeList as List, areEqual } from 'react-window';
import TansformerSelect from './TransformerSelect';

// If list items are expensive to render,
// Consider using PureComponent to avoid unnecessary re-renders.
// https://reactjs.org/docs/react-api.html#reactpurecomponent
const Row = memo(function Row({ data, index, style }) {
  // Data passed to List as "itemData" is available as props.data
  const { items, onSelect, transformers } = data;
  const item = items[index];

  return (
    <div style={style} className="border-t">
      <div className="grid grid-cols-5 gap-4 items-center p-2">
        <div className="flex flex-row truncate ">
          <Checkbox
            id="select"
            onClick={() => onSelect(index)}
            checked={item.isSelected}
            type="button"
            className="self-center mr-4"
          />
          <Cell value={item.schema} />
        </div>
        <Cell value={item.table} />
        <Cell value={item.column} />
        <Cell value={item.dataType} />
        <div className=" ">
          <FormField
            name={`mappings.${index}.transformer.value`}
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <div className="flex flex-row space-x-2  ">
                    <div className="w-[175px]">
                      <TansformerSelect
                        transformers={transformers || []}
                        value={field.value}
                        onSelect={field.onChange}
                        placeholder="Search transformers..."
                        defaultValue="passthrough"
                      />
                    </div>
                    <EditTransformerOptions
                      transformer={transformers?.find(
                        (item) => item.value == field.value
                      )}
                      index={index}
                    />
                  </div>
                </FormControl>
              </FormItem>
            )}
          />
        </div>
      </div>
    </div>
  );
}, areEqual);
Row.displayName = 'row';

// This helper function memoizes incoming props,
// To avoid causing unnecessary re-renders pure Row components.
// This is only needed since we are passing multiple props with a wrapper object.
// If we were only passing a single, stable value (e.g. items),
// We could just pass the value directly.
const createRowData = memoize((items, onSelect, onSelectAll, transformers) => ({
  items,
  onSelect,
  onSelectAll,
  transformers,
}));

// In this example, "items" is an Array of objects to render,
// and "onSelect" is a function that updates an item's state.
function Example({
  height,
  rows,
  onSelect,
  onSelectAll,
  width,
  transformers,
  bulkSelect,
  setBulkSelect,
}) {
  // Bundle additional data to list items using the "itemData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  const rowData = createRowData(rows, onSelect, onSelectAll, transformers);

  return (
    <div className={`space-y-4 border rounded-md w-[${width}px]`}>
      <div className={`grid grid-cols-5 gap-2 pl-2 pt-4`}>
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
          Schema
        </div>
        <div className="">Table</div>
        <div className="">Column</div>
        <div className="">Data Type</div>
        <div className="">Transformer</div>
      </div>
      <List
        height={height}
        itemCount={rows.length}
        itemData={rowData}
        itemSize={50}
        width={width}
      >
        {Row}
      </List>
    </div>
  );
}

interface SchemaListProps {
  data: Row[];
  transformers?: Transformer[];
  width: number;
  height: number;
}

export const TableList = memo(function TableList({
  data,
  transformers,
  width,
  height,
}: SchemaListProps) {
  const [rows, setRows] = useState(data);
  const [transformer, setTransformer] = useState<string>('');
  const [bulkSelect, setBulkSelect] = useState(false);
  const form = useFormContext();

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

  return (
    <div className={`space-y-4 w-[${width + 20}px]`}>
      <div className="flex items-center justify-between">
        <div className="w-[250px]">
          <TansformerSelect
            transformers={transformers || []}
            value={transformer}
            onSelect={(value) => {
              rows.forEach((r, index) => {
                if (r.isSelected) {
                  form.setValue(`mappings.${index}.transformer.value`, value, {
                    shouldDirty: true,
                  });
                }
              });
              onSelectAll(false);
              setBulkSelect(false);
              setTransformer('');
            }}
            placeholder="Bulk update transformers..."
          />
        </div>
        <Button
          variant="outline"
          type="button"
          // onClick={() => {
          //   table.setColumnFilters([]);
          //   onClearFilters();
          // }}
        >
          Clear filters
          <UpdateIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </div>

      <Example
        height={height}
        rows={rows}
        onSelect={onSelect}
        onSelectAll={onSelectAll}
        transformers={transformers}
        bulkSelect={bulkSelect}
        setBulkSelect={setBulkSelect}
        width={width}
      />
    </div>
  );
});

interface CellProps {
  value: string;
}

function Cell(props: CellProps) {
  const { value } = props;
  return <span className="truncate font-medium text-sm">{value}</span>;
}
