'use client';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { ReactElement } from 'react';
import JobHaltNewAdditionSwitch from './components/JobHaltNewAdditionSwitch';
import SourceConnectionCard from './components/SourceConnectionCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, mutate } = useGetJob(id);
  return (
    <div className="job-details-container">
      <PageHeader
        header="Source Connection"
        description="Manage job's source connection"
      />

      <div className="space-y-10">
        <SourceConnectionCard jobId={id} />
        <Card>
          <CardHeader>
            <CardTitle>Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <JobHaltNewAdditionSwitch
              isHalted={
                data?.job?.sourceOptions?.haltOnNewColumnAddition || false
              }
              mutate={mutate}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
