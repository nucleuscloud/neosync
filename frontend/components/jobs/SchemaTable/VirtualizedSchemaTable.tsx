import EditTransformerOptions, {
  handleTransformerMetadata,
} from '@/app/transformers/EditTransformerOptions';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command';
import { FormControl, FormField, FormItem } from '@/components/ui/form';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/libs/utils';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { CaretSortIcon, CheckIcon, UpdateIcon } from '@radix-ui/react-icons';
import memoize from 'memoize-one';
import { memo, useCallback, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { FixedSizeList as List, areEqual } from 'react-window';

// If list items are expensive to render,
// Consider using PureComponent to avoid unnecessary re-renders.
// https://reactjs.org/docs/react-api.html#reactpurecomponent
const Row = memo(function Row({ data, index, style }) {
  // Data passed to List as "itemData" is available as props.data
  const { items, toggleItemActive, transformers } = data;
  const item = items[index];

  return (
    <div style={style} className="border-t">
      <div className="grid grid-cols-5 gap-4 items-center p-2">
        <div className="flex flex-row truncate ">
          <Checkbox
            id="select"
            onClick={() => toggleItemActive(index)}
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
const createItemData = memoize(
  (items, toggleItemActive, toggleAllItemActive, transformers) => ({
    items,
    toggleItemActive,
    toggleAllItemActive,
    transformers,
  })
);

// In this example, "items" is an Array of objects to render,
// and "toggleItemActive" is a function that updates an item's state.
function Example({
  height,
  items,
  toggleItemActive,
  toggleAllItemActive,
  width,
  transformers,
  bulkSelect,
  setBulkSelect,
}) {
  // Bundle additional data to list items using the "itemData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  const itemData = createItemData(
    items,
    toggleItemActive,
    toggleAllItemActive,
    transformers
  );

  return (
    <div className={`space-y-4 border rounded-md p-4 w-[${width + 20}px]`}>
      <div className={`grid grid-cols-5 gap-2`}>
        <div className="">
          <div className="flex flex-row">
            <Checkbox
              id="select"
              onClick={() => {
                toggleAllItemActive(!bulkSelect);
                setBulkSelect(!bulkSelect);
              }}
              checked={bulkSelect}
              type="button"
              className="self-center mr-4"
            />
            Schema
          </div>
        </div>
        <div className="">Table</div>
        <div className="">Column</div>
        <div className="">Data Type</div>
        <div className="">Transformer</div>
      </div>
      <List
        height={height}
        itemCount={items.length}
        itemData={itemData}
        itemSize={50}
        width={width}
      >
        {Row}
      </List>
    </div>
  );
}

interface Row {
  isSelected: boolean;
  table: string;
  transformer: {
    value: string;
    config: {};
  };
  schema: string;
  column: string;
  dataType: string;
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
  const [items, setItems] = useState(data);
  const [transformer, setTransformer] = useState<string>('');
  const [bulkSelect, setBulkSelect] = useState(false);
  const form = useFormContext();

  const toggleItemActive = useCallback((index: number) => {
    setItems((prevItems) => {
      const newItems = [...prevItems];
      newItems[index] = {
        ...newItems[index],
        isSelected: !newItems[index].isSelected,
      };
      return newItems;
    });
  }, []);

  const toggleAllItemActive = useCallback((isSelected: boolean) => {
    setItems((prevItems) => {
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
        <div className="flex flex-1 items-center space-x-2">
          <BulkTansformerSelect
            transformers={transformers || []}
            value={transformer}
            onSelect={(value) => {
              items.forEach((r, index) => {
                if (r.isSelected) {
                  form.setValue(`mappings.${index}.transformer.value`, value, {
                    shouldDirty: true,
                  });
                }
              });
              toggleAllItemActive(false);
              setBulkSelect(false);
              setTransformer('');
            }}
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
        items={items}
        toggleItemActive={toggleItemActive}
        toggleAllItemActive={toggleAllItemActive}
        transformers={transformers}
        bulkSelect={bulkSelect}
        setBulkSelect={setBulkSelect}
        width={width}
      />
    </div>
  );
});

interface TransformersSelectProps {
  transformers: Transformer[];
  value: string;
  onSelect: (value: string) => void;
}

function TansformerSelect(props: TransformersSelectProps) {
  const { transformers, value, onSelect } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="justify-between w-full" //whitespace-nowrap
        >
          <div className="truncate overflow-hidden text-ellipsis whitespace-nowrap">
            {value
              ? transformers.find((t) => t.value === value)?.value
              : 'Transformer'}
          </div>
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className=" p-0">
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
                defaultValue={'passthrough'}
              >
                <CheckIcon
                  className={cn(
                    'mr-2 h-4 w-4',
                    value == t.value ? 'opacity-100' : 'opacity-0'
                  )}
                />
                {handleTransformerMetadata(t).name}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
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
            ? transformers.find((t) => t.value === value)?.value
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
                {t.value}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

interface CellProps {
  value: string;
}

function Cell(props: CellProps) {
  const { value } = props;

  return <span className="truncate font-medium text-sm">{value}</span>;
}
