'use client';
import { PageProps } from '@/components/types';

import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import Spinner from '@/components/Spinner';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
import { TiCancel } from 'react-icons/ti';

import ResourceId from '@/components/ResourceId';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import {
  refreshEventsWhenEventsIncomplete,
  refreshJobRunWhenJobRunning,
} from '@/libs/utils';
import { formatDateTime, getErrorMessage } from '@/util/util';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  cancelJobRun,
  deleteJobRun,
  getJobRun,
  getJobRunEvents,
  terminateJobRun,
} from '@neosync/sdk/connectquery';
import { ArrowRightIcon, Cross2Icon, TrashIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import JobRunStatus from '../components/JobRunStatus';
import JobRunActivityTable from './components/JobRunActivityTable';
import JobRunLogs from './components/JobRunLogs';

export default function Page({ params }: PageProps): ReactElement {
  const { account } = useAccount();
  const accountId = account?.id || '';
  const id = params?.id ?? '';
  const router = useRouter();
  const { toast } = useToast();
  const { data: systemAppConfigData, isLoading: isSystemAppConfigDataLoading } =
    useGetSystemAppConfig();
  const {
    data,
    isLoading,
    refetch: mutate,
  } = useQuery(
    getJobRun,
    { jobRunId: id, accountId: accountId },
    {
      enabled: !!id && !!accountId,
      refetchInterval(query) {
        return query.state.data
          ? refreshJobRunWhenJobRunning(query.state.data)
          : 0;
      },
    }
  );
  const jobRun = data?.jobRun;

  const {
    data: eventData,
    isLoading: eventsIsLoading,
    isFetching: isValidating,
    refetch: eventMutate,
  } = useQuery(
    getJobRunEvents,
    { jobRunId: id, accountId: accountId },
    {
      enabled: !!id && !!accountId,
      refetchInterval(query) {
        return query.state.data
          ? refreshEventsWhenEventsIncomplete(query.state.data)
          : 0;
      },
    }
  );

  const { mutateAsync: removeJobRunAsync } = useMutation(deleteJobRun);
  const { mutateAsync: cancelJobRunAsync } = useMutation(cancelJobRun);
  const { mutateAsync: terminateJobRunAsync } = useMutation(terminateJobRun);

  async function onDelete(): Promise<void> {
    try {
      await removeJobRunAsync({ accountId: accountId, jobRunId: id });
      toast({
        title: 'Job run removed successfully!',
      });
      router.push(`/${account?.name}/runs`);
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove job run',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function onCancel(): Promise<void> {
    try {
      await cancelJobRunAsync({ accountId, jobRunId: id });
      toast({
        title: 'Job run canceled successfully!',
      });
      mutate();
      eventMutate();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to cancel job run',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function onTerminate(): Promise<void> {
    try {
      await terminateJobRunAsync({ accountId, jobRunId: id });
      toast({
        title: 'Job run terminated successfully!',
      });
      mutate();
      eventMutate();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to terminate job run',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Job Run Details"
          extraHeading={
            <div className="flex flex-row space-x-4">
              <DeleteConfirmationDialog
                trigger={
                  <Button variant="destructive">
                    <ButtonText leftIcon={<TrashIcon />} text="Delete" />
                  </Button>
                }
                headerText="Are you sure you want to delete this job run?"
                description=""
                onConfirm={async () => onDelete()}
              />
              {(jobRun?.status === JobRunStatusEnum.RUNNING ||
                jobRun?.status === JobRunStatusEnum.PENDING) && (
                <div className="flex flex-row gap-4">
                  <ConfirmationDialog
                    trigger={
                      <Button variant="default">
                        <ButtonText leftIcon={<Cross2Icon />} text="Cancel" />
                      </Button>
                    }
                    headerText="Are you sure you want to cancel this job run?"
                    description=""
                    onConfirm={async () => onCancel()}
                    buttonText="Cancel"
                    buttonVariant="default"
                    buttonIcon={<Cross2Icon />}
                  />
                  <ConfirmationDialog
                    trigger={
                      <Button>
                        <ButtonText leftIcon={<TiCancel />} text="Terminate" />
                      </Button>
                    }
                    headerText="Are you sure you want to terminate this job run?"
                    description=""
                    onConfirm={async () => onTerminate()}
                    buttonText="Terminate"
                    buttonVariant="default"
                    buttonIcon={<Cross2Icon />}
                  />
                </div>
              )}
              <ButtonLink jobId={jobRun?.jobId} />
            </div>
          }
          subHeadings={
            <ResourceId
              labelText={jobRun?.id ?? ''}
              copyText={jobRun?.id ?? ''}
              onHoverText="Copy the Run ID"
            />
          }
        />
      }
      containerClassName="runs-page"
    >
      {isLoading ? (
        <div className="space-y-24">
          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4`}
          >
            <Skeleton className="w-full h-24 rounded-lg" />
            <Skeleton className="w-full h-24 rounded-lg" />
            <Skeleton className="w-full h-24 rounded-lg" />
            <Skeleton className="w-full h-24 rounded-lg" />
          </div>

          <SkeletonTable />
        </div>
      ) : (
        <div className="space-y-12">
          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4`}
          >
            <StatCard
              header="Status"
              content={
                <JobRunStatus
                  status={jobRun?.status}
                  containerClassName="px-0"
                  badgeClassName="text-lg"
                />
              }
            />
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
          {!isSystemAppConfigDataLoading &&
            systemAppConfigData?.enableRunLogs && (
              <div>
                <JobRunLogs accountId={accountId} runId={id} />
              </div>
            )}
          <div className="space-y-4">
            <div className="flex flex-row items-center space-x-2">
              <h2 className="text-2xl font-bold tracking-tight">Activity</h2>
              {isValidating && <Spinner />}
            </div>
            {eventsIsLoading ? (
              <SkeletonTable />
            ) : (
              <JobRunActivityTable jobRunEvents={eventData?.events} />
            )}
          </div>
        </div>
      )}
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
  const seconds = Math.round(differenceValue % 60);
  if (minutes === 0) {
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
  const { account } = useAccount();
  if (!props.jobId) {
    return <></>;
  }
  return (
    <Button
      variant="outline"
      onClick={() => router.push(`/${account?.name}/jobs/${props.jobId}`)}
    >
      <ButtonText
        text="View Job"
        rightIcon={<ArrowRightIcon className="ml-2 h-4 w-4" />}
      />
    </Button>
  );
}
