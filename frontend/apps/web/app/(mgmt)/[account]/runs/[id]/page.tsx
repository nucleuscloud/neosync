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
import { refreshWhenJobRunning, useGetJobRun } from '@/libs/hooks/useGetJobRun';
import { JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
import { TiCancel } from 'react-icons/ti';

import { CopyButton } from '@/components/CopyButton';
import {
  refreshEventsWhenEventsIncomplete,
  useGetJobRunEvents,
} from '@/libs/hooks/useGetJobRunEvents';
import {
  refreshLogsWhenRunNotComplete,
  useGetJobRunLogs,
} from '@/libs/hooks/useGetJobRunLogs';
import { formatDateTime, getErrorMessage } from '@/util/util';
import {
  ArrowRightIcon,
  Cross2Icon,
  ReloadIcon,
  TrashIcon,
} from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useRef, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import { VariableSizeList as List } from 'react-window';
import JobRunStatus from '../components/JobRunStatus';
import JobRunActivityTable from './components/JobRunActivityTable';

type WindowSize = {
  width: number;
  height: number;
};

export default function Page({ params }: PageProps): ReactElement {
  const { account } = useAccount();
  const accountId = account?.id || '';
  const id = params?.id ?? '';
  const router = useRouter();
  const { toast } = useToast();
  const { data, isLoading, mutate } = useGetJobRun(id, accountId, {
    refreshIntervalFn: refreshWhenJobRunning,
  });

  const {
    data: eventData,
    isLoading: eventsIsLoading,
    isValidating,
    mutate: eventMutate,
  } = useGetJobRunEvents(id, accountId, {
    refreshIntervalFn: refreshEventsWhenEventsIncomplete,
  });
  // TODO fix refresh
  const {
    data: logsData,
    isLoading: isLogsLoading,
    isValidating: isLogsValidating,
    mutate: logsMutate,
  } = useGetJobRunLogs(id, accountId, {
    refreshIntervalFn: refreshLogsWhenRunNotComplete,
  });

  const logs = logsData || [];
  const jobRun = data?.jobRun;

  const [windowSize, setWindowSize] = useState<WindowSize>({
    width: window.innerWidth,
    height: window.innerHeight,
  });
  const listRef = useRef<List<string[]> | null>(null);

  useEffect(() => {
    function handleResize() {
      setWindowSize({
        width: window.innerWidth,
        height: window.innerHeight,
      });
      if (listRef.current) {
        listRef.current.resetAfterIndex(0);
      }
    }

    window.addEventListener('resize', handleResize);
    handleResize();
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  async function onDelete(): Promise<void> {
    try {
      await removeJobRun(id, accountId);
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
      await cancelJobRun(id, accountId);
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
      await terminateJobRun(id, accountId);
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

  function onRefreshClick(): void {
    logsMutate();
  }

  function getLogLineSize(index: number): number {
    const log = logs[index];
    const maxLineWidth = windowSize.width;
    const estimatedLineWidth = log.length * 10;
    const numberOfLines = Math.ceil(estimatedLineWidth / maxLineWidth);
    const height = 5 + numberOfLines * 28;
    return height;
  }

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Job Run Details"
          description={jobRun?.id || ''}
          copyIcon={
            <CopyButton
              onHoverText="Copy the Run ID"
              textToCopy={jobRun?.id || ''}
              onCopiedText="Success!"
              buttonVariant="outline"
            />
          }
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
                <JobRunStatus status={jobRun?.status} className="text-lg" />
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
            {logs?.some((l) => l.includes('ERROR')) && (
              <AlertDestructive
                key="log-error"
                title="Log Errors"
                description="check the logs for errors"
              />
            )}
          </div>
          <div className="space-y-4">
            <div className="flex flex-row items-center space-x-2">
              <h1 className="text-2xl font-bold tracking-tight">Logs</h1>
              <Button
                className={isLogsValidating ? 'animate-spin' : ''}
                disabled={isLogsValidating}
                variant="ghost"
                size="icon"
                onClick={() => onRefreshClick()}
              >
                <ReloadIcon className="h-4 w-4" />
              </Button>
            </div>
            {isLogsLoading ? (
              <SkeletonTable />
            ) : (
              <div className="h-[500px] w-full p-4">
                <AutoSizer>
                  {({ height, width }) => (
                    <List
                      ref={listRef}
                      className="border rounded-md dark:border-gray-700"
                      height={height}
                      itemCount={logs.length}
                      itemSize={getLogLineSize}
                      width={width}
                      itemKey={(index: number) => logs[index]}
                      itemData={logs}
                    >
                      {({ index, style }) => {
                        return (
                          <p className="p-2" style={style}>
                            {logs[index]}
                          </p>
                        );
                      }}
                    </List>
                  )}
                </AutoSizer>
              </div>
            )}
          </div>
          <div className="space-y-4">
            <div className="flex flex-row items-center space-x-2">
              <h1 className="text-2xl font-bold tracking-tight">Activity</h1>
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

async function removeJobRun(
  jobRunId: string,
  accountId: string
): Promise<void> {
  const res = await fetch(`/api/accounts/${accountId}/runs/${jobRunId}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

async function cancelJobRun(
  jobRunId: string,
  accountId: string
): Promise<void> {
  const res = await fetch(
    `/api/accounts/${accountId}/runs/${jobRunId}/cancel`,
    {
      method: 'PUT',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

async function terminateJobRun(
  jobRunId: string,
  accountId: string
): Promise<void> {
  const res = await fetch(
    `/api/accounts/${accountId}/runs/${jobRunId}/terminate`,
    {
      method: 'PUT',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
