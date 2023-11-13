import Link from '@docusaurus/Link';
import React from 'react';
import { Badge } from '../components/Badge';

interface Props {
  title: string;
  type: string;
  apiRef: string;
}

export const TransformerPageHeader = (props: Props) => {
  const { title, type, apiRef } = props;
  return (
    <div className="transformer-page-header">
      <div>{title}</div>
      <Badge>{type}</Badge>
      <Link href={apiRef} className="flex items-center">
        <Badge>API reference</Badge>
      </Link>
    </div>
  );
};
