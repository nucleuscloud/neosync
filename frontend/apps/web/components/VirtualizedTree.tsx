import { Checkbox } from '@/components/ui/checkbox';
import { ChevronDownIcon, ChevronUpIcon } from '@radix-ui/react-icons';
import memoizeOne from 'memoize-one';
import { CSSProperties, memo, useCallback, useEffect, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List, areEqual } from 'react-window';

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
  onNodeSelect: (id: string, isSelected: boolean) => void;
}

export const VirtualizedTree = ({ data, onNodeSelect }: TreeProps) => {
  const [closedNodeIds, setClosedNodeIds] = useState<string[]>([]);
  const [nodes, setNodes] = useState(() => flattenOpened(data, closedNodeIds));

  useEffect(() => {
    const newFlattenedData = flattenOpened(data, closedNodeIds);
    setNodes(newFlattenedData);
  }, [data, closedNodeIds]);

  const onClose = useCallback(
    (node: Node) => {
      setClosedNodeIds((prevIds) => {
        const newIds = node.collapsed
          ? prevIds.filter((id) => id !== node.id)
          : [...prevIds, node.id];

        return newIds;
      });
      setNodes(flattenOpened(data, closedNodeIds));
    },
    [data, closedNodeIds]
  );

  const onSelect = useCallback((index: number, node: Node) => {
    setNodes((prevNodes) => {
      const newNodes = [...prevNodes];
      newNodes[index] = {
        ...newNodes[index],
        isSelected: !newNodes[index].isSelected,
      };
      const isSelected = !node.isSelected;

      // handle selecting children nodes
      var areSibilingNodesSelected = false;
      var depth = node.depth + 1;
      var i = index + 1;
      while (depth != 1) {
        const nextItem = newNodes[i];
        if (!nextItem) {
          break;
        }
        if (nextItem.depth != 1 && node.hasChildren) {
          newNodes[i] = {
            ...nextItem,
            isSelected: isSelected,
          };
        }
        if (nextItem.depth === node.depth && nextItem.isSelected) {
          areSibilingNodesSelected = true;
        }
        depth = nextItem.depth;
        i = i + 1;
      }

      // handle selecting parent nodes
      var depth = node.depth;
      var i = index - 1;
      while (depth != 1) {
        const nextItem = newNodes[i];
        if (nextItem.depth != depth) {
          newNodes[i] = {
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
      return newNodes;
    });
  }, []);

  const nodeData = getNodeData(onClose, onSelect, nodes, onNodeSelect);

  return (
    <AutoSizer>
      {({ height, width }) => (
        <List
          className="border rounded-md dark:border-gray-700"
          height={height}
          itemCount={nodes.length}
          itemSize={38}
          width={width}
          itemKey={(index: number) => nodes[index].id}
          itemData={nodeData}
        >
          {Node}
        </List>
      )}
    </AutoSizer>
  );
};

interface NodeProps {
  index: number;
  style: CSSProperties;
  data: {
    nodes: Node[];
    onSelect: (index: number, node: Node) => void;
    onNodeSelect: (id: string, isSelected: boolean) => void;
    onClose: (node: Node) => void;
  };
}

const Node = memo(({ data, index, style }: NodeProps) => {
  const { nodes, onClose, onSelect, onNodeSelect } = data;
  const node = nodes[index];
  const left = node.depth != 1 && `pl-${node.depth * 4}`;
  const hover = node.hasChildren && `hover:bg-muted p-2`;
  return (
    <div style={style}>
      <div className={`flex flex-row ${left} ${hover}`}>
        <Checkbox
          id={node.id}
          onClick={() => {
            onSelect(index, node);
            onNodeSelect(node.id, !node.isSelected);
          }}
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
            onClose(node);
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

const getNodeData = memoizeOne((onClose, onSelect, nodes, onNodeSelect) => ({
  onClose,
  onSelect,
  nodes,
  onNodeSelect,
}));

const flattenOpened = (data: TreeData[], closedNodeIds: string[]) => {
  const result: Node[] = [];
  for (let node of data) {
    flattenNode(node, 1, result, closedNodeIds);
  }
  return result;
};

const flattenNode = (
  node: TreeData,
  depth: number,
  result: Node[],
  closedNodeIds: string[]
) => {
  const { id, name, children, isSelected } = node;

  let collapsed = closedNodeIds.includes(id);
  result.push({
    id,
    name,
    hasChildren: children && children.length > 0,
    depth,
    collapsed,
    isSelected,
    isIndeterminate: false,
  });

  if (!collapsed && children) {
    for (let child of children) {
      flattenNode(child, depth + 1, result, closedNodeIds);
    }
  }
};
