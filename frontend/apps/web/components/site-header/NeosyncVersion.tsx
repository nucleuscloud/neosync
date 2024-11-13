'use client';
import { useQuery } from '@connectrpc/connect-query';
import { getSystemInformation } from '@neosync/sdk/connectquery';
import { ReactElement } from 'react';

export default function NeosyncVersion(): ReactElement | null {
  const { data } = useQuery(getSystemInformation);
  if (!data?.version) {
    return null;
  }
  return <p className="text-sm tracking-tight">{data.version}</p>;
}
