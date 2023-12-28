import Link from '@docusaurus/Link';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import React, { ReactElement } from 'react';
import { Card } from './CustomCardList';

export function CustomCard(props: Card): ReactElement {
  const { title, description, icon, link } = props;
  return (
    <Link
      href={link}
      className="border border-gray-200 dark:border-gray-900 dark:bg-[#232323] hover:dark:border-gray-700 hover:border-[#acc7f9] hover:shadow-md hover:text-gray-800 text-gray-900 dark:text-gray-300 rounded-xl p-4 flex flex-col justify-between space-y-2 dark:hover:text-gray-100 no-underline hover:no-underline group"
    >
      <div className="flex flex-col space-y-4">
        <div className="flex flex-row gap-2 items-center">
          <div className="border-2 border-blue-200 dark:border-gray-800 rounded-lg bg-blue-100 dark:bg-gray-700 dark:text-gray-300 p-2">
            {icon}
          </div>
          <div className="font-semibold text-base">{title}</div>
        </div>
        <div className="font-light">{description}</div>
      </div>
      <div className="flex justify-end transition-transform duration-300 transform group-hover:translate-x-[4px]">
        <ArrowRightIcon />
      </div>
    </Link>
  );
}
