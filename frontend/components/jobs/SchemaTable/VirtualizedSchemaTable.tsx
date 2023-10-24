import { Checkbox } from '@/components/ui/checkbox';
import memoize from 'memoize-one';
import { memo, useCallback, useState } from 'react';
import { FixedSizeList as List, areEqual } from 'react-window';

// If list items are expensive to render,
// Consider using PureComponent to avoid unnecessary re-renders.
// https://reactjs.org/docs/react-api.html#reactpurecomponent
const Row = memo(({ data, index, style }) => {
  // Data passed to List as "itemData" is available as props.data
  const { items, toggleItemActive } = data;
  const item = items[index];

  return (
    <div style={style}>
      <div className="grid grid-cols-11 gap-2">
        <div className="col-span-1">
          <Checkbox
            id="select"
            onClick={() => toggleItemActive(index)}
            checked={item.isSelected}
            type="button"
          />
        </div>
        <div className="col-span-2 truncate">{item.schema}</div>
        <div className="col-span-2 truncate">{item.table}</div>
        <div className="col-span-2 truncate">{item.column}</div>
        <div className="col-span-2 truncate">{item.dataType}</div>
        <div className="col-span-2 truncate">{item.transformer.value}</div>
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
  (items, toggleItemActive, toggleAllItemActive) => ({
    items,
    toggleItemActive,
    toggleAllItemActive,
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
}) {
  // Bundle additional data to list items using the "itemData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  const itemData = createItemData(items, toggleItemActive, toggleAllItemActive);
  const [allToggled, setAllToggled] = useState(false);

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-11 gap-2">
        <div className="col-span-1">
          <Checkbox
            id="select"
            onClick={() => {
              toggleAllItemActive(!allToggled);
              setAllToggled(!allToggled);
            }}
            checked={allToggled}
            type="button"
          />
        </div>
        <div className="col-span-2">Schema</div>
        <div className="col-span-2">Table</div>
        <div className="col-span-2">Column</div>
        <div className="col-span-2">Data Type</div>
        <div className="col-span-2">Transformer</div>
      </div>
      <List
        height={height}
        itemCount={items.length}
        itemData={itemData}
        itemSize={45}
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
}

export const TableList = memo(function TableList({ data }: SchemaListProps) {
  const [items, setItems] = useState(data);

  // useEffect(() => {
  //   setItems(data);
  // }, []);
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
    console.log('isSelected', isSelected);
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
    <Example
      height={700}
      items={items}
      toggleItemActive={toggleItemActive}
      toggleAllItemActive={toggleAllItemActive}
      width={1300}
    />
  );
});
