import { ThemeClassNames } from '@docusaurus/theme-common';
import type { Props } from '@theme/DocSidebarItem/Html';
import clsx from 'clsx';
import React from 'react';

import styles from './styles.module.css';

export default function DocSidebarItemHtml({
  item,
  level,
  index,
}: Props): JSX.Element {
  const { value, defaultStyle, className } = item;
  return (
    <li
      className={clsx(
        ThemeClassNames.docs.docSidebarItemLink,
        ThemeClassNames.docs.docSidebarItemLinkLevel(level),
        defaultStyle && [styles.menuHtmlItem, 'menu__list-item'],
        className
      )}
      key={index}
      // eslint-disable-next-line react/no-danger
      dangerouslySetInnerHTML={{ __html: value }}
    />
  );
}
