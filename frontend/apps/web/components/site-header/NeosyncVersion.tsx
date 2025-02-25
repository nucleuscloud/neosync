'use client';
import { useQuery } from '@connectrpc/connect-query';
import { UserAccountService } from '@neosync/sdk';
import { ReactElement } from 'react';

export default function NeosyncVersion(): ReactElement | null {
  const { data } = useQuery(UserAccountService.method.getSystemInformation);
  if (!data?.version) {
    return null;
  }
  return <p className="text-sm tracking-tight">{data.version}</p>;
}
