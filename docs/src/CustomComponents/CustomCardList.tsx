import React, { ReactElement } from 'react';
import { CustomCard } from './CustomCard';

interface Props {
  cards: Card[];
}

export interface Card {
  title: string;
  description: string;
  icon: JSX.Element;
  link: string;
}

export function CustomCardList(props: Props): ReactElement {
  const { cards } = props;
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-2 no-underline">
      {cards.map((item) => (
        <CustomCard
          key={item.title}
          title={item.title}
          description={item.description}
          icon={item.icon}
          link={item.link}
        />
      ))}
    </div>
  );
}
