import { Checkbox } from '@/components/ui/checkbox';
import memoize from 'memoize-one';
import { CSSProperties, memo, useCallback, useState } from 'react';
import { FixedSizeList as List, areEqual } from 'react-window';

interface Row {
  isSelected: boolean;
  id: string;
  name: string;
  children?: Row[];
}

// If list items are expensive to render,
// Consider using PureComponent to avoid unnecessary re-renders.
// https://reactjs.org/docs/react-api.html#reactpurecomponent
interface RowProps {
  index: number;
  style: CSSProperties;
  data: {
    rows: Row[];
    onSelect: (index: number) => void;
    onSelectAll: (value: boolean) => void;
  };
}

// If list items are expensive to render,
// Consider using PureComponent to avoid unnecessary re-renders.
// https://reactjs.org/docs/react-api.html#reactpurecomponent
const Row = memo(function Row({ data, index, style }: RowProps) {
  // Data passed to List as "itemData" is available as props.data
  const { rows, onSelect } = data;
  const row = rows[index];

  return (
    <div style={style}>
      <div className="flex flex-row truncate ">
        <Checkbox
          id="select"
          onClick={() => onSelect(index)}
          checked={row.isSelected}
          type="button"
          className="self-center mr-4"
        />
        <span className="truncate font-medium text-sm">{row.name}</span>
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
const createRowData = memoize(
  (rows: Row[], onSelect: (index: number) => void) => ({
    rows,
    onSelect,
  })
);

// In this example, "items" is an Array of objects to render,
// and "toggleItemActive" is a function that updates an item's state.
function Example({ height, items, onSelect, width }) {
  // Bundle additional data to list items using the "itemData" prop.
  // It will be accessible to item renderers as props.data.
  // Memoize this data to avoid bypassing shouldComponentUpdate().
  const itemData = createRowData(items, onSelect);

  return (
    <div className="border rounded-md p-4">
      <List
        height={height}
        itemCount={items.length}
        itemData={itemData}
        itemSize={35}
        width={width}
      >
        {Row}
      </List>
    </div>
  );
}

interface SchemaTreeProps {
  data: Row[];
  width: number;
  height: number;
}

export const SchemaTree = memo(function SchemaTree({
  data,
  height,
  width,
}: SchemaTreeProps) {
  const [items, setItems] = useState(data);

  const onSelect = useCallback((index) => {
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
    <Example height={height} items={items} onSelect={onSelect} width={width} />
  );
});
