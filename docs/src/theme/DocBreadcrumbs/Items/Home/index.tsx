import Link from '@docusaurus/Link';
import { translate } from '@docusaurus/Translate';
import useBaseUrl from '@docusaurus/useBaseUrl';
import IconHome from '@theme/Icon/Home';
import React, { ReactElement } from 'react';

import styles from './styles.module.css';

interface Props {}
export default function HomeBreadcrumbItem(props: Props): ReactElement {
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
        <IconHome className={styles['home-icon']} />
      </Link>
    </li>
  );
}
