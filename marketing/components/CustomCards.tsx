import { ReactElement } from 'react';
import { Card, CustomCard } from './CustomCard';

interface Props {
  cards: Card[];
}

export function CustomCards(props: Props): ReactElement {
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
