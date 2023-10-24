import React, { useState } from 'react';
import { FixedSizeList as List } from 'react-window';

interface TreeNode {
  id: string;
  name: string;
  children?: TreeNode[];
}

const Tree: React.FC<{ node: TreeNode }> = ({ node }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isChecked, setIsChecked] = useState(false);

  return (
    <div style={{ marginLeft: 20 }}>
      <div>
        <input
          type="checkbox"
          checked={isChecked}
          onChange={() => setIsChecked(!isChecked)}
        />
        {/* {node.children && (
          <button type="button" onClick={() => setIsOpen(!isOpen)}>
            {isOpen ? '-' : '+'}
          </button>
        )} */}
        {node.name}
      </div>
      {node.children && (
        <div>
          {node.children.map((child) => (
            // <Tree key={child.id} node={child} />
            <input
              type="checkbox"
              key={child.id}
              checked={isChecked}
              onChange={() => setIsChecked(!isChecked)}
            />
          ))}
        </div>
      )}
    </div>
  );
};

type TreeProps = {
  data: TreeNode[];
};

const TreeList = ({ data }: TreeProps) => {
  const renderRow = ({ index, style }) => (
    <div style={style}>
      <Tree node={data[index]} />
    </div>
  );

  return (
    <List height={700} itemCount={data.length} itemSize={30} width={400}>
      {renderRow}
    </List>
  );
};

export default TreeList;
