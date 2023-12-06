import Link from '@docusaurus/Link';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import React, { ReactElement } from 'react';
import { Card } from './CustomCardList';

export function CustomCard(props: Card): ReactElement {
  const { title, description, icon, link } = props;
  return (
    <Link
      href={link}
      className="custom-card no-underline hover:no-underline group text-gray-900"
    >
      <div className="flex flex-col space-y-4">
        <div className="flex flex-row gap-2 items-center">
          <div className="border-2 border-blue-200 rounded-lg bg-blue-100 p-2">
            {icon}
          </div>
          <div className="card-title">{title}</div>
        </div>
        <div className="card-text">{description}</div>
      </div>
      <div className="flex justify-end transition-transform duration-300 transform group-hover:translate-x-[4px]">
        <ArrowRightIcon />
      </div>
    </Link>
  );
}
