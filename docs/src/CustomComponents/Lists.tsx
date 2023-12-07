import React from 'react';
import { BsDot } from 'react-icons/bs';

interface Items {
  items: string[];
}

export const OrderedListComponent = ({ items }: Items) => {
  const numberWidth = `${String(items.length).length}em`;

  return (
    <div className="pb-8">
      {items.map((item, index) => (
        <div key={item} className="flex flex-row items-start gap-2">
          <div style={{ width: numberWidth, textAlign: 'left' }}>
            {index + 1}.
          </div>
          <div>{item}</div>
        </div>
      ))}
    </div>
  );
};

export const UnorderedListComponent = ({ items }: Items) => {
  const numberWidth = `${String(items.length).length}em`;
  return (
    <div className="pb-8">
      {items.map((item, index) => (
        <div key={item} className="flex flex-row items-start gap-2">
          <div style={{ width: numberWidth, textAlign: 'left' }}>
            <BsDot />
          </div>
          <div>{item}</div>
        </div>
      ))}
    </div>
  );
};
