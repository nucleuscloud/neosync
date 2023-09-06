'use client';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { ReactElement } from 'react';
import SourceConnectionCard from './components/SourceConnectionCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data } = useGetJob(id);
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
            <div className="w-96">
              <div className="flex flex-row items-center justify-between rounded-lg border p-4">
                <div className="space-y-0.5">
                  <Label className="text-base">
                    Halt Job on new column addition
                  </Label>
                  <p className="text-sm text-muted-foreground">
                    Stops job runs if new column is detected
                  </p>
                </div>
                <Switch
                  checked={data?.job?.sourceOptions?.haltOnNewColumnAddition}
                  onCheckedChange={() => {}}
                />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
