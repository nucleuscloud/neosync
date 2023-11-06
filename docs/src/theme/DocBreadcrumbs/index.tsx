/**
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

import Link from '@docusaurus/Link';
import { translate } from '@docusaurus/Translate';
import { ThemeClassNames } from '@docusaurus/theme-common';
import {
  useHomePageRoute,
  useSidebarBreadcrumbs,
} from '@docusaurus/theme-common/internal';
import clsx from 'clsx';
import React, { type ReactNode } from 'react';

import { ChevronRightIcon, HomeIcon } from '@radix-ui/react-icons';
import styles from './styles.module.css';

// TODO move to design system folder
function BreadcrumbsItemLink({
  children,
  href,
  isLast,
}: {
  children: ReactNode;
  href: string | undefined;
  isLast: boolean;
}): JSX.Element {
  const className = 'breadcrumbs__link';
  if (isLast) {
    return (
      <span className={className} itemProp="name">
        {children}
      </span>
    );
  }
  return href ? (
    <Link className={className} href={href} itemProp="item">
      <span itemProp="name">{children}</span>
    </Link>
  ) : (
    // TODO Google search console doesn't like breadcrumb items without href.
    // The schema doesn't seem to require `id` for each `item`, although Google
    // insist to infer one, even if it's invalid. Removing `itemProp="item
    // name"` for now, since I don't know how to properly fix it.
    // See https://github.com/facebook/docusaurus/issues/7241
    <span className={className}>{children}</span>
  );
}

// TODO move to design system folder
function BreadcrumbsItem({
  children,
  active,
  index,
  addMicrodata,
}: {
  children: ReactNode;
  active?: boolean;
  index: number;
  addMicrodata: boolean;
}): JSX.Element {
  return (
    <li
      {...(addMicrodata && {
        itemScope: true,
        itemProp: 'itemListElement',
        itemType: 'https://schema.org/ListItem',
      })}
      className={clsx(
        'breadcrumbs__item',
        {
          'breadcrumbs__item--active': active,
        },
        ''
      )}
    >
      {children}
      <meta itemProp="position" content={String(index + 1)} />
    </li>
  );
}

export default function DocBreadcrumbs(): JSX.Element | null {
  const breadcrumbs = useSidebarBreadcrumbs();
  const homePageRoute = useHomePageRoute();

  if (!breadcrumbs) {
    return null;
  }

  return (
    <nav
      className={clsx(
        ThemeClassNames.docs.docBreadcrumbs,
        styles.breadcrumbsContainer
      )}
      aria-label={translate({
        id: 'theme.docs.breadcrumbs.navAriaLabel',
        message: 'Breadcrumbs',
        description: 'The ARIA label for the breadcrumbs',
      })}
    >
      <ul
        className="breadcrumbs"
        itemScope
        itemType="https://schema.org/BreadcrumbList"
      >
        <div className="flex flex-row gap-2">
          {homePageRoute && (
            <LocalHomeBreadcrumbItem homePageRoute={homePageRoute} />
          )}
          {breadcrumbs.map((item, idx) => {
            const isLast = idx === breadcrumbs.length - 1;
            const href =
              item.type === 'category' && item.linkUnlisted
                ? undefined
                : item.href;
            return (
              <BreadcrumbsItem
                key={idx}
                active={isLast}
                index={idx}
                addMicrodata={!!href}
              >
                <BreadcrumbsItemLink href={href} isLast={isLast}>
                  {item.label}
                </BreadcrumbsItemLink>
              </BreadcrumbsItem>
            );
          })}
        </div>
      </ul>
    </nav>
  );
}

interface LocalHomeBreadcrumbItemsProps {
  homePageRoute: any;
}
function LocalHomeBreadcrumbItem(props: LocalHomeBreadcrumbItemsProps) {
  const { homePageRoute } = props;
  return (
    <div className="flex flex-row items-center space-x-2">
      <Link href={homePageRoute.path}>
        <HomeIcon className="home-icon" />
      </Link>
      <ChevronRightIcon className="dark:text-gray-100 text-gray-800" />
    </div>
  );
}
