import Link from '@docusaurus/Link';
import { translate } from '@docusaurus/Translate';
import useBaseUrl from '@docusaurus/useBaseUrl';
import React, { ReactElement } from 'react';

import { HomeIcon } from '@radix-ui/react-icons';
import styles from './styles.module.css';

interface Props {}
export default function HomeBreadcrumbItem(props: Props): ReactElement {
  const homeHref = useBaseUrl('/');

  return (
    <li className="breadcrumbs__item breadcrumbs__home">
      <Link
        aria-label={translate({
          id: 'theme.docs.breadcrumbs.home',
          message: 'Home page',
          description: 'The ARIA label for the home page in the breadcrumbs',
        })}
        className="breadcrumbs__link pl-0"
        href={homeHref}
      >
        <HomeIcon className={styles['home-icon']} />
      </Link>
    </li>
  );
}
