import Link from '@docusaurus/Link';
import {
  useHomePageRoute,
  useSidebarBreadcrumbs,
} from '@docusaurus/theme-common/internal';
import { ChevronRightIcon, HomeIcon } from '@radix-ui/react-icons';
import React from 'react';

export default function DocBreadcrumbs() {
  const breadcrumbs = useSidebarBreadcrumbs();
  const homePageRoute = useHomePageRoute();

  if (!breadcrumbs) {
    return null;
  }

  return (
    <div className="inline-flex flex-row items-center mb-10 rounded-full text-gray-800 p-2 space-x-2">
      <div className="flex flex-row items-center space-x-2">
        <Link href={homePageRoute.path}>
          <HomeIcon className="home-icon" />
        </Link>
        <ChevronRightIcon className="dark:text-gray-100 text-gray-800" />
      </div>
      <div className="flex flex-row items-center space-x-2">
        {breadcrumbs.map((item, idx) => {
          const isLast = idx === breadcrumbs.length - 1;
          if (isLast) {
            return (
              <div
                className="breadcrumb-label"
                itemProp="name"
                key={item.label}
              >
                {item.label}
              </div>
            );
          } else {
            return (
              <div
                className="cursor-pointer flex-row items-center flex space-x-2"
                key={item.label}
              >
                <span itemProp="name" className="breadcrumb-label">
                  {item.label}
                </span>
                <ChevronRightIcon className="dark:text-gray-100 text-gray-800" />
              </div>
            );
          }
        })}
      </div>
    </div>
  );
}
