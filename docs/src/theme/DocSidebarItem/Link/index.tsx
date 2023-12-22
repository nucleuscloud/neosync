import Link from '@docusaurus/Link';
import isInternalUrl from '@docusaurus/isInternalUrl';
import { ThemeClassNames } from '@docusaurus/theme-common';
import { isActiveSidebarItem } from '@docusaurus/theme-common/internal';
import { IconHandler } from '@site/src/CustomComponents/IconHandler';
import { cn } from '@site/src/utils';
import IconExternalLink from '@theme/Icon/ExternalLink';
import clsx from 'clsx';
import React from 'react';
import styles from './styles.module.css';

export default function DocSidebarItemLink({
  item,
  onItemClick,
  activePath,
  level,
  index,
  ...props
}) {
  const { href, label, className, autoAddBaseUrl } = item;
  const isActive = isActiveSidebarItem(item, activePath);
  const isInternalLink = isInternalUrl(href);

  return (
    <li
      className={clsx(
        ThemeClassNames.docs.docSidebarItemLink,
        ThemeClassNames.docs.docSidebarItemLinkLevel(level),
        'menu__list-item',
        className
      )}
      key={label}
    >
      <Link
        className={clsx(
          'menu__link',
          !isInternalLink && styles.menuExternalLink,
          {
            'menu__link--active': isActive,
          }
        )}
        autoAddBaseUrl={autoAddBaseUrl}
        aria-current={isActive ? 'page' : undefined}
        to={href}
        {...(isInternalLink && {
          onClick: onItemClick ? () => onItemClick(item) : undefined,
        })}
        {...props}
      >
        <div className="menu__link--item">
          <div
            className={cn(
              isActive ? 'text-blue-500' : 'text-gray-900 dark:text-gray-300'
            )}
          >
            {IconHandler(item.label)}
          </div>
          {label}
          {!isInternalLink && <IconExternalLink />}
        </div>
      </Link>
    </li>
  );
}
