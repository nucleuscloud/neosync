import { ArrowRightIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

export interface Card {
  title: string;
  description: string;
  icon: JSX.Element;
  link: string;
}

interface Props extends Card {}

export function CustomCard(props: Props): ReactElement {
  const { title, description, icon, link } = props;
  return (
    <Link
      href={link}
      className="custom-card border border-gray-200 dark:border-gray-900 dark:bg-[#232323] hover:dark:border-gray-700 hover:border-[#acc7f9] hover:shadow-md hover:text-gray-800  dark:text-gray-300 rounded-xl p-4 flex flex-col justify-between space-y-2 dark:hover:text-gray-100  no-underline hover:no-underline group text-gray-900"
    >
      <div className="flex flex-col space-y-4">
        <div className="flex flex-row gap-2 items-center">
          <div className="custom-card-icon border-2 border-blue-200 dark:border-gray-800 rounded-lg bg-blue-100 dark:bg-gray-700 dark:text-gray-300 p-2">
            {icon}
          </div>
          <div className="card-title font-semibold text-base">{title}</div>
        </div>
        <div className="card-text font-light">{description}</div>
      </div>
      <div className="flex justify-end transition-transform duration-300 transform group-hover:translate-x-[4px]">
        <ArrowRightIcon />
      </div>
    </Link>
  );
}
