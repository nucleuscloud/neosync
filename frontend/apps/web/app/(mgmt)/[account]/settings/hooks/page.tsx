'use client';

import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { ReactElement } from 'react';
import HooksCard from './components/HooksCard';

export default function Page(): ReactElement<any> {
  const { account } = useAccount();
  const { data: configData, isLoading } = useGetSystemAppConfig();
  if (isLoading || !account?.id) {
    return <Skeleton className="w-full h-12" />;
  }
  if (!configData?.isAccountHooksEnabled) {
    return <HooksDisabledAlert />;
  }
  return (
    <div className="job-hooks-page-container">
      <HooksCard accountId={account.id} />
    </div>
  );
}

function HooksDisabledAlert(): ReactElement<any> {
  return (
    <div>
      <Alert variant="warning">
        <AlertTitle>Account Hooks are not currently enabled</AlertTitle>
        <AlertDescription>
          To enable them, please update Neosync configuration or contact your
          system administrator.
        </AlertDescription>
      </Alert>
    </div>
  );
}
