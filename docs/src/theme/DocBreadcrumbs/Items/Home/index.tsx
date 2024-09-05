import Link from '@docusaurus/Link';
import { translate } from '@docusaurus/Translate';
import useBaseUrl from '@docusaurus/useBaseUrl';
import React from 'react';

import { HomeIcon } from '@radix-ui/react-icons';
import styles from './styles.module.css';

export default function HomeBreadcrumbItem(): JSX.Element {
  const homeHref = useBaseUrl('/');

  return (
    <li className="breadcrumbs__item">
      <Link
        aria-label={translate({
          id: 'theme.docs.breadcrumbs.home',
          message: 'Home page',
          description: 'The ARIA label for the home page in the breadcrumbs',
        })}
        className="breadcrumbs__link"
        href={homeHref}
      >
        <HomeIcon className={styles.breadcrumbHomeIcon} />
      </Link>
    </li>
  );
}
