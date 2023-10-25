import { Checkbox } from '@/components/ui/checkbox';
import { ChevronDownIcon, ChevronUpIcon } from '@radix-ui/react-icons';
import memoizeOne from 'memoize-one';
import { memo, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List, areEqual } from 'react-window';

const Row = memo(({ data, index, style }) => {
  const { flattenedData, onOpen, onSelect } = data;
  const node = flattenedData[index];
  const left = node.depth != 1 && `pl-${node.depth * 4}`;
  const hover = node.hasChildren && `hover:bg-muted p-2`;
  return (
    <div style={style}>
      <div className={`flex flex-row ${left} ${hover}`}>
        <Checkbox
          id="select"
          onClick={() => onSelect(index)}
          checked={node.isSelected}
          type="button"
          className="self-center mr-2"
        />

        <div
          className="flex flex-row w-full justify-between items-center"
          onClick={() => {
            onOpen(node);
          }}
        >
          <span className="truncate font-medium text-sm">{node.name}</span>
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
    });

    if (!collapsed && children) {
      for (let child of children) {
        flattenNode(child, depth + 1, result);
      }
    }
  };

  const onOpen = (node) =>
    node.collapsed
      ? setOpenedNodeIds([...openedNodeIds, node.id])
      : setOpenedNodeIds(openedNodeIds.filter((id) => id !== node.id));

  const onSelect = (e, node) => {
    //e.stopPropagation();
  };

  const flattenedData = flattenOpened(data);

  const itemData = getItemData(onOpen, onSelect, flattenedData);

  return (
    <AutoSizer>
      {({ height, width }) => (
        <List
          className="List"
          height={height}
          itemCount={flattenedData.length}
          itemSize={38}
          width={width}
          itemKey={(index) => flattenedData[index].id}
          itemData={itemData}
        >
          {Row}
        </List>
      )}
    </AutoSizer>
  );
};
