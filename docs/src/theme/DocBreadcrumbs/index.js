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
      <div>
        {breadcrumbs.map((item, idx) => {
          const isLast = idx === breadcrumbs.length - 1;
          if (isLast) {
            return (
              <span
                className="cursor-pointer bg-gray-800 text-sm text-gray-100 p-2 rounded-full hover:bg-gray-900"
                itemProp="name"
              >
                {item.label}
              </span>
            );
          } else {
            return (
              <div className=" cursor-pointer">
                <Link href={item.href} itemProp="item">
                  <span itemProp="name">{item.label}</span>
                </Link>
              </div>
            );
          }
        })}
      </div>
    </div>
  );
}
