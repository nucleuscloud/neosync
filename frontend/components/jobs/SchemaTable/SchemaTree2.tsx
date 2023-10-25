import { Checkbox } from '@/components/ui/checkbox';
import memoizeOne from 'memoize-one';
import { memo, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List, areEqual } from 'react-window';

const Row = memo(({ data, index, style }) => {
  const { flattenedData, onOpen, onSelect } = data;
  const node = flattenedData[index];
  const left = node.depth * 20;
  return (
    <div className="item-background" style={style}>
      <div className="flex flex-row ">
        <Checkbox
          id="select"
          onClick={() => onSelect(index)}
          checked={node.isSelected}
          type="button"
          className="self-center mr-4"
        />

        <div
          className={`${node.hasChildren ? 'tree-branch' : ''} ${
            node.collapsed ? 'tree-item-closed' : 'tree-item-open'
          }`}
          // onClick={(e) => onSelect(e, node)}
          onClick={() => onOpen(node)}
          style={{
            position: 'absolute',
            left: `${left}px`,
            width: `calc(100% - ${left}px)`,
          }}
        >
          <span className="truncate font-medium text-sm">{node.name}</span>
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
          itemSize={32}
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
