'use client';

import { PageProps } from '@/components/types';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { ReactElement } from 'react';
import HooksCard from './components/HooksCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data: configData, isLoading } = useGetSystemAppConfig();
  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }
  if (!configData?.isJobHooksEnabled) {
    return <HooksDisabledAlert />;
  }
  return (
    <div className="job-hooks-page-container">
      <HooksCard jobId={id} />
    </div>
  );
}

function HooksDisabledAlert(): ReactElement {
  return (
    <div>
      <Alert variant="warning">
        <AlertTitle>Job Hooks are not currently enabled</AlertTitle>
        <AlertDescription>
          To enable them, please update Neosync configuration or contact your
          system administrator.
        </AlertDescription>
      </Alert>
    </div>
  );
}
