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
    // <div onClick={() => toggleItemActive(index)} style={style}>
    //   {`${item.schema} ${item.table} ${item.column} ${item.dataType} ${item.transformer.value}`}
    // </div>
    <div className="grid grid-cols-6 gap-4">
      <div>
        <Checkbox
          id={`${item.schema}-${item.table}-${item.column}`}
          onClick={() => toggleItemActive(index)}
          checked={item.isSelected}
          type="button"
        />
      </div>
      <div>{item.schema}</div>
      <div>{item.table}</div>
      <div>{item.column}</div>
      <div>{item.dataType}</div>
      <div>{item.transformer.value}</div>
    </div>
  );
}, areEqual);

Row.displayName = 'row';

// This helper function memoizes incoming props,
// To avoid causing unnecessary re-renders pure Row components.
// This is only needed since we are passing multiple props with a wrapper object.
// If we were only passing a single, stable value (e.g. items),
// We could just pass the value directly.
const createItemData = memoize((items, toggleItemActive) => ({
  items,
  toggleItemActive,
}));

// In this example, "items" is an Array of objects to render,
// and "toggleItemActive" is a function that updates an item's state.
function Example({ height, items, toggleItemActive, width }) {
  // Bundle additional data to list items using the "itemData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  const itemData = createItemData(items, toggleItemActive);

  return (
    <List
      height={height}
      itemCount={items.length}
      itemData={itemData}
      itemSize={35}
      width={width}
    >
      {Row}
    </List>
  );
}

interface TableRow {
  transformer: {
    value: string;
    config: {};
  };
  schema: string;
  table: string;
  column: string;
  dataType: string;
  isSelected: boolean;
}

interface TableProps {
  data: TableRow[];
}

export const TableList = memo(function TableList({ data }: TableProps) {
  const [items, setItems] = useState(data);

  const toggleItemActive = useCallback((index) => {
    setItems((prevItems) => {
      const newItems = [...prevItems];
      newItems[index] = {
        ...newItems[index],
        isSelected: !newItems[index].isSelected,
      };
      return newItems;
    });
  }, []);

  return (
    <Example
      height={700}
      items={items}
      toggleItemActive={toggleItemActive}
      width={1300}
    />
  );
});
