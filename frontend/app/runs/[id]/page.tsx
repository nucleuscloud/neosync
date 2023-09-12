'use client';
import { PageProps } from '@/components/types';

import ProgressNav from '@/components/Progress';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetJobRun } from '@/libs/hooks/useGetJobRun';
import { JobRunStatusType } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTime } from '@/util/util';
import { ReactElement } from 'react';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetJobRun(id);

  const jobRun = data?.jobRun;

  if (isLoading) {
    return <Skeleton />;
  }

  return (
    <OverviewContainer
      Header={
        <PageHeader header="Job Run Details" description={jobRun?.name || ''} />
      }
      containerClassName="runs-page"
    >
      <div className="space-y-8">
        <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4`}>
          <StatCard
            header="Status"
            content={getStatusBadge(jobRun?.status?.status)}
          />
          <StatCard
            header="Start Time"
            content={formatDateTime(jobRun?.status?.startTime?.toDate())}
          />
          <StatCard
            header="End Time"
            content={formatDateTime(jobRun?.status?.completionTime?.toDate())}
          />
          <StatCard
            header="Duration"
            content={getDuration(
              jobRun?.status?.completionTime?.toDate(),
              jobRun?.status?.startTime?.toDate()
            )}
          />
        </div>
        <ProgressNav />
      </div>
    </OverviewContainer>
  );
}

interface StatCardProps {
  header: string;
  content?: JSX.Element | string;
}

function StatCard(props: StatCardProps): ReactElement {
  const { header, content } = props;
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{header}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-lg font-bold">{content}</div>
      </CardContent>
    </Card>
  );
}

function getDuration(dateTimeValue2?: Date, dateTimeValue1?: Date): string {
  if (!dateTimeValue1 || !dateTimeValue2) {
    return '';
  }
  var differenceValue =
    (dateTimeValue2.getTime() - dateTimeValue1.getTime()) / 1000;
  const minutes = Math.abs(Math.round(differenceValue / 60));
  const seconds = differenceValue % 60;
  if (minutes == 0) {
    return `${seconds} seconds`;
  }
  return `${minutes} minutes ${seconds} seconds`;
}

function getStatusBadge(status?: number): JSX.Element {
  if (!status) {
    return <Badge variant="outline">Unknown</Badge>;
  }
  switch (status) {
    case JobRunStatusType.JOB_RUN_STATUS_COMPLETE:
      return <Badge className="bg-green-600">Complete</Badge>;
    case JobRunStatusType.JOB_RUN_STATUS_ERROR:
      return <Badge variant="destructive">Failed</Badge>;
    case JobRunStatusType.JOB_RUN_STATUS_RUNNING:
      return <Badge className="bg-yellow-600">Running</Badge>;
    default:
      return <Badge variant="outline">Unknown</Badge>;
  }
}
