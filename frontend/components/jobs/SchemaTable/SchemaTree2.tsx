import { Checkbox } from '@/components/ui/checkbox';
import { ChevronDownIcon, ChevronUpIcon } from '@radix-ui/react-icons';
import memoizeOne from 'memoize-one';
import { memo, useCallback, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List, areEqual } from 'react-window';
const test = [
  {
    id: 'neosync',
    name: 'neosync',
    hasChildren: true,
    depth: 1,
    collapsed: false,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.regions',
    name: 'regions',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.countries',
    name: 'countries',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.locations',
    name: 'locations',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.jobs',
    name: 'jobs',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.departments',
    name: 'departments',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.employees',
    name: 'employees',
    depth: 2,
    hasChildren: true,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.employees.other',
    name: 'other',
    depth: 3,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'neosync.employees.thing',
    name: 'thing',
    depth: 3,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'nucleus',
    name: 'nucleus',
    hasChildren: true,
    depth: 1,
    collapsed: false,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'nucleus.regions',
    name: 'regions',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'nucleus.countries',
    name: 'countries',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'nucleus.locations',
    name: 'locations',
    depth: 2,
    collapsed: true,
    isSelected: false,
    isIndeterminate: false,
  },
  {
    id: 'mancity',
    name: 'man city',
    depth: 1,
    collapsed: false,
    isSelected: false,
    isIndeterminate: false,
  },
];

const Row = memo(({ data, index, style }) => {
  const { flattenedData, onOpen, onSelect } = data;
  const node = flattenedData[index];
  const left = node.depth != 1 && `pl-${node.depth * 4}`;
  const hover = node.hasChildren && `hover:bg-muted p-2`;
  return (
    <div style={style}>
      <div className={`flex flex-row ${left} ${hover}`}>
        <Checkbox
          id={node.id}
          onClick={() => onSelect(index, node)}
          checked={node.isSelected}
          indeterminate={node.isIndeterminate}
          type="button"
          className="self-center mr-2"
        />
        <label
          htmlFor={node.id}
          className={`text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 border`}
        >
          {node.name} - {node.depth}
        </label>

        <div
          className="flex flex-row items-center justify-end grow border"
          onClick={() => {
            onOpen(node);
          }}
        >
          <div className="mr-2">
            {node.hasChildren && node.collapsed && <ChevronDownIcon />}
            {node.hasChildren && !node.collapsed && <ChevronUpIcon />}
          </div>
        </div>
      </div>
    </div>
  );
}, areEqual);
Row.displayName = 'row';

const getItemData = memoizeOne((onOpen, onSelect, flattenedData) => ({
  onOpen,
  onSelect,
  flattenedData,
}));

interface Row {
  isSelected: boolean;
  id: string;
  name: string;
  children?: Row[];
}

interface Node {
  id: string;
  name: string;
  hasChildren: boolean;
  depth: number;
  collapsed: boolean;
  isSelected: boolean;
  isIndeterminate: boolean;
}

interface SchemaTreeProps {
  data: Row[];
}

export const SchemaTreeAutoResize = ({ data }: SchemaTreeProps) => {
  const [openedNodeIds, setOpenedNodeIds] = useState(data.map((d) => d.id));

  const flattenOpened = (treeData: Row[]) => {
    const result = [];
    for (let node of data) {
      flattenNode(node, 1, result);
    }
    return result;
  };

  const flattenNode = (node, depth, result) => {
    const { id, name, children } = node;
    let collapsed = !openedNodeIds.includes(id);
    result.push({
      id,
      name,
      hasChildren: children && children.length > 0,
      depth,
      collapsed,
      isSelected: node.isSelected,
      isIndeterminate: false,
    });

    if (!collapsed && children) {
      for (let child of children) {
        flattenNode(child, depth + 1, result);
      }
    }
  };

  const onOpen = (node) => {
    console.log(
      'node',
      JSON.stringify(node),
      'openedNodeIds',
      openedNodeIds,
      'new',
      openedNodeIds.filter((id) => id !== node.id)
    );
    node.collapsed
      ? setOpenedNodeIds([...openedNodeIds, node.id])
      : setOpenedNodeIds(openedNodeIds.filter((id) => id !== node.id));
  };

  const flattenedData = flattenOpened(data);

  const [items, setItems] = useState(flattenedData);

  const onSelect = useCallback((index: number, node: Node) => {
    setItems((prevItems) => {
      const newItems = [...prevItems];
      newItems[index] = {
        ...newItems[index],
        isSelected: !newItems[index].isSelected,
      };

      // handle selecting parent nodes
      const isSelected = !node.isSelected;
      var depth = node.depth;
      var i = index - 1;
      while (depth != 1) {
        const nextItem = newItems[i];
        if (nextItem.depth != depth) {
          newItems[i] = {
            ...nextItem,
            isIndeterminate: isSelected,
            isSelected: isSelected,
          };
          depth = nextItem.depth;
        }
        i = i - 1;
      }

      // handle selecting children nodes
      if (node.hasChildren) {
        console.log('node', JSON.stringify(node));
        var depth = node.depth + 1;
        var i = index + 1;
        while (depth != 1) {
          const nextItem = newItems[i];
          console.log('index', i, 'nextItem', JSON.stringify(nextItem));
          if (!nextItem) {
            break;
          }

          if (nextItem.depth != 1) {
            newItems[i] = {
              ...nextItem,
              isSelected: isSelected,
            };
          }
          depth = nextItem.depth;
          i = i + 1;
        }
      }
      return newItems;
    });
  }, []);

  const itemData = getItemData(onOpen, onSelect, items);

  return (
    <AutoSizer>
      {({ height, width }) => (
        <List
          className="border rounded-md"
          height={height}
          itemCount={items.length}
          itemSize={38}
          width={width}
          itemKey={(index) => items[index].id}
          itemData={itemData}
        >
          {Row}
        </List>
      )}
    </AutoSizer>
  );
};
