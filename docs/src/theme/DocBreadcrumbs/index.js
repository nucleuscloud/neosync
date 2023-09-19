import React from "react";
import {
  useSidebarBreadcrumbs,
  useHomePageRoute,
} from "@docusaurus/theme-common/internal";
import Link from "@docusaurus/Link";
import { ChevronRightIcon, HomeIcon } from "@radix-ui/react-icons";

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
          <HomeIcon className="dark:text-gray-100 text-gray-800" />
        </Link>
        <ChevronRightIcon className="dark:text-gray-100 text-gray-800" />
      </div>
      <div className="flex flex-row items-center">
        {breadcrumbs.map((item, idx) => {
          const isLast = idx === breadcrumbs.length - 1;
          if (isLast) {
            return (
              <div className="flex flex-row items-center" key={item.docId}>
                <span
                  className="cursor-pointer text-sm text-gray-100 p-2 rounded-full"
                  itemProp="name"
                  key={item.label}
                >
                  {item.label}
                </span>
              </div>
            );
          } else {
            return (
              <div
                className="cursor-pointer flex-row items-center flex"
                key={item.label}
              >
                <Link href={item.href} itemProp="item">
                  <span
                    itemProp="name"
                    className="cursor-pointer text-sm text-gray-100 p-2 rounded-full"
                  >
                    {item.label}
                  </span>
                </Link>
                <ChevronRightIcon className="dark:text-gray-100 text-gray-800" />
              </div>
            );
          }
        })}
      </div>
    </div>
  );
}
