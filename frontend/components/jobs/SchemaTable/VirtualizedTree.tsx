import { Checkbox } from '@/components/ui/checkbox';
import { ChevronDownIcon, ChevronUpIcon } from '@radix-ui/react-icons';
import memoizeOne from 'memoize-one';
import { CSSProperties, memo, useCallback, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List, areEqual } from 'react-window';

interface RowProps {
  index: number;
  style: CSSProperties;
  data: {
    flattenedData: Node[];
    onSelect: (index: number, node: Node) => void;
    onOpen: (node: Node) => void;
  };
}

const Node = memo(({ data, index, style }: RowProps) => {
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
          className={`text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70`}
        >
          {node.name}
        </label>

        <div
          className="flex flex-row items-center justify-end grow"
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
Node.displayName = 'node';

const getItemData = memoizeOne((onOpen, onSelect, flattenedData) => ({
  onOpen,
  onSelect,
  flattenedData,
}));

interface TreeData {
  isSelected: boolean;
  id: string;
  name: string;
  children?: TreeData[];
}

interface Node {
  id: string;
  name: string;
  hasChildren?: boolean;
  depth: number;
  collapsed: boolean;
  isSelected?: boolean;
  isIndeterminate?: boolean;
}

interface TreeProps {
  data: TreeData[];
}

export const VirtualizedTree = ({ data }: TreeProps) => {
  const [openedNodeIds, setOpenedNodeIds] = useState(data.map((d) => d.id));

  const onOpen = useCallback((node: Node) => {
    setOpenedNodeIds((prevIds) => {
      const newIds = node.collapsed
        ? [...prevIds, node.id]
        : prevIds.filter((id) => id !== node.id);

      setItems(flattenOpened(data, newIds));
      return newIds;
    });
  }, []);

  const flattenedData = flattenOpened(data, openedNodeIds);

  const [items, setItems] = useState(flattenedData);

  const onSelect = useCallback((index: number, node: Node) => {
    setItems((prevItems) => {
      const newItems = [...prevItems];
      newItems[index] = {
        ...newItems[index],
        isSelected: !newItems[index].isSelected,
      };
      const isSelected = !node.isSelected;

      // handle selecting children nodes
      var areSibilingNodesSelected = false;
      var depth = node.depth + 1;
      var i = index + 1;
      while (depth != 1) {
        const nextItem = newItems[i];
        if (!nextItem) {
          break;
        }
        if (nextItem.depth != 1 && node.hasChildren) {
          newItems[i] = {
            ...nextItem,
            isSelected: isSelected,
          };
        }
        if (nextItem.depth == node.depth && nextItem.isSelected) {
          areSibilingNodesSelected = true;
        }
        depth = nextItem.depth;
        i = i + 1;
      }

      // handle selecting parent nodes
      var depth = node.depth;
      var i = index - 1;
      while (depth != 1) {
        const nextItem = newItems[i];
        if (nextItem.depth != depth) {
          newItems[i] = {
            ...nextItem,
            isIndeterminate: areSibilingNodesSelected || isSelected,
            isSelected: areSibilingNodesSelected || isSelected,
          };
          depth = nextItem.depth;
        } else {
          if (nextItem.isSelected) {
            areSibilingNodesSelected = true;
          }
        }
        i = i - 1;
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
          itemKey={(index: number) => items[index].id}
          itemData={itemData}
        >
          {Node}
        </List>
      )}
    </AutoSizer>
  );
};

const flattenOpened = (data: TreeData[], openNodeIds: string[]) => {
  const result: Node[] = [];
  for (let node of data) {
    flattenNode(node, 1, result, openNodeIds);
  }
  return result;
};

const flattenNode = (
  node: TreeData,
  depth: number,
  result: Node[],
  openNodeIds: string[]
) => {
  const { id, name, children } = node;
  let collapsed = !openNodeIds.includes(id);
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
      flattenNode(child, depth + 1, result, openNodeIds);
    }
  }
};
