'use client';
import { PageProps } from '@/components/types';

import ProgressNav from '@/components/Progress';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetJobRun } from '@/libs/hooks/useGetJobRun';
import { useGetJobRunEvents } from '@/libs/hooks/useGetJobRunEvents';
import { formatDateTime } from '@/util/util';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { JOB_RUN_STATUS } from '../components/status';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetJobRun(id);
  const { data: jobRunEvents, isLoading: jobRunEventsLoading } =
    useGetJobRunEvents(id);

  const jobRun = data?.jobRun;
  const events = jobRunEvents?.events || [];

  const progressNavItems = events.map((e) => {
    return {
      title: `${e.name} - ${e.type}`,
      description: formatDateTime(e.createdAt?.toDate()) || '',
    };
  });

  if (isLoading) {
    return <Skeleton />;
  }
  const status = JOB_RUN_STATUS.find(
    (status) => status.value === jobRun?.status
  );

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Job Run Details"
          description={jobRun?.name || ''}
          extraHeading={<ButtonLink jobId={jobRun?.jobId} />}
        />
      }
      containerClassName="runs-page"
    >
      <div className="space-y-8">
        <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4`}>
          <StatCard header="Status" content={status?.badge} />
          <StatCard
            header="Start Time"
            content={formatDateTime(jobRun?.startedAt?.toDate())}
          />
          <StatCard
            header="Completion Time"
            content={formatDateTime(jobRun?.completedAt?.toDate())}
          />
          <StatCard
            header="Duration"
            content={getDuration(
              jobRun?.completedAt?.toDate(),
              jobRun?.startedAt?.toDate()
            )}
          />
        </div>
        <div className="space-y-4">
          {jobRun?.pendingActivities.map((a) => {
            if (a.lastFailure) {
              return (
                <AlertDestructive
                  key={a.activityName}
                  title={a.activityName}
                  description={a.lastFailure?.message || ''}
                />
              );
            }
          })}
        </div>
        <div className="space-y-4">
          <h1 className="text-xl font-semibold tracking-tight">Steps</h1>
          {jobRunEventsLoading ? (
            <Skeleton />
          ) : (
            <ProgressNav items={progressNavItems} />
          )}
        </div>
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

interface AlertProps {
  title: string;
  description: string;
}

function AlertDestructive(props: AlertProps): ReactElement {
  return (
    <Alert variant="destructive">
      <AlertTitle>{`${props.title}: ${props.description}`}</AlertTitle>
    </Alert>
  );
}

interface ButtonProps {
  jobId?: string;
}

function ButtonLink(props: ButtonProps): ReactElement {
  const router = useRouter();
  if (!props.jobId) {
    return <></>;
  }
  return (
    <Button
      variant="outline"
      onClick={() => router.push(`/jobs/${props.jobId}`)}
    >
      View Job
      <ArrowRightIcon className="ml-2 h-4 w-4" />
    </Button>
  );
}
