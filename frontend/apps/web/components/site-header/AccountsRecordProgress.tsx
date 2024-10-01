'use client';
import { UserAccountType } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useAccount } from '../providers/account-provider';
import RecordsProgressBar from './RecordsProgressBar';

export default function AccountsRecordProgress(): ReactElement | null {
  const { account } = useAccount();
  const idtype = 'accountId';

  if (account?.type === UserAccountType.PERSONAL) {
    return <RecordsProgressBar identifier={account.id ?? ''} idtype={idtype} />;
  }

  return null;
}
