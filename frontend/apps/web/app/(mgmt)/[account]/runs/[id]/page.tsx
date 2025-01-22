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
import { JobRunStatus as JobRunStatusEnum, JobService } from '@neosync/sdk';
import { TiCancel } from 'react-icons/ti';

import { CopyButton } from '@/components/CopyButton';
import ResourceId from '@/components/ResourceId';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import {
  refreshEventsWhenEventsIncomplete,
  refreshJobRunWhenJobRunning,
} from '@/libs/utils';
import { formatDateTime, getErrorMessage } from '@/util/util';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { Editor } from '@monaco-editor/react';
import { ArrowRightIcon, Cross2Icon, TrashIcon } from '@radix-ui/react-icons';
import { formatDuration, intervalToDuration } from 'date-fns';
import { useTheme } from 'next-themes';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import { format as formatSql } from 'sql-formatter';
import yaml from 'yaml';
import JobRunStatus from '../components/JobRunStatus';
import JobRunActivityErrors from './components/JobRunActivityErrors';
import JobRunActivityTable from './components/JobRunActivityTable';
import JobRunLogs from './components/JobRunLogs';

export default function Page({ params }: PageProps): ReactElement {
  const { account } = useAccount();
  const accountId = account?.id || '';
  const id = decodeURIComponent(params?.id ?? '');
  const router = useRouter();
  const { data: systemAppConfigData, isLoading: isSystemAppConfigDataLoading } =
    useGetSystemAppConfig();
  const {
    data,
    isLoading,
    refetch: mutate,
  } = useQuery(
    JobService.method.getJobRun,
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
    JobService.method.getJobRunEvents,
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

  const { mutateAsync: removeJobRunAsync } = useMutation(
    JobService.method.deleteJobRun
  );
  const { mutateAsync: cancelJobRunAsync } = useMutation(
    JobService.method.cancelJobRun
  );
  const { mutateAsync: terminateJobRunAsync } = useMutation(
    JobService.method.terminateJobRun
  );
  const { mutateAsync: getRunContextAsync } = useMutation(
    JobService.method.getRunContext
  );

  const [isViewSelectDialogOpen, setIsSelectDialogOpen] =
    useState<boolean>(false);
  const [activeSelectQuery, setActiveSelectQuery] = useState<SelectQuery>({
    schema: '',
    table: '',
    select: '',
  });
  const [isRetrievingRunContext, setIsRetrievingRunContext] =
    useState<boolean>(false);

  const [duration, setDuration] = useState<string>('');

  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (jobRun?.startedAt && jobRun?.status === JobRunStatusEnum.RUNNING) {
      const updateDuration = () => {
        if (jobRun?.startedAt) {
          setDuration(getDuration(new Date(), timestampDate(jobRun.startedAt)));
        }
      };

      updateDuration();
      // sets up an interval to call the timer every second
      timer = setInterval(updateDuration, 1000);
    } else if (jobRun?.completedAt && jobRun?.startedAt) {
      setDuration(
        getDuration(
          timestampDate(jobRun.completedAt),
          timestampDate(jobRun.startedAt)
        )
      );
    }
    // cleans up and restarts the interval if the job isn't done yet
    return () => {
      if (timer) clearInterval(timer);
    };
  }, [jobRun?.startedAt, jobRun?.completedAt, jobRun?.status]);

  async function onDelete(): Promise<void> {
    try {
      await removeJobRunAsync({ accountId: accountId, jobRunId: id });
      toast.success('Job run removed successfully!');
      router.push(`/${account?.name}/runs`);
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onCancel(): Promise<void> {
    try {
      await cancelJobRunAsync({ accountId, jobRunId: id });
      toast.success('Job run canceled successfully!');
      mutate();
      eventMutate();
    } catch (err) {
      console.error(err);
      toast.error('Unable to cancel job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onTerminate(): Promise<void> {
    try {
      await terminateJobRunAsync({ accountId, jobRunId: id });
      toast.success('Job run terminated successfully!');
      mutate();
      eventMutate();
    } catch (err) {
      console.error(err);
      toast.error('Unable to terminate job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onViewSelectClicked(
    schema: string,
    table: string
  ): Promise<void> {
    if (isRetrievingRunContext) {
      return;
    }
    setIsRetrievingRunContext(true);
    try {
      const rcResp = await getRunContextAsync({
        id: {
          accountId: accountId,
          externalId: buildBenthosRunCtxExternalId(schema, table),
          jobRunId: id,
        },
      });
      const runCtx = parseUint8ArrayToYaml(rcResp.value);
      if (isValidRunContext(runCtx)) {
        setActiveSelectQuery({
          schema,
          table,
          select: runCtx.input.pooled_sql_raw.query,
        });
        setIsSelectDialogOpen(true);
      } else {
        toast.error('Unable to parse run context', {
          description: `Was unable to pull Select Query out of run context for ${schema}.${table}`,
        });
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to retrieve select for table', {
        description: getErrorMessage(err),
      });
    } finally {
      setIsRetrievingRunContext(false);
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
                headerText="Delete Job Run?"
                description="Are you sure you want to delete this job run?"
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
                    headerText="Cancel Job Run?"
                    description="Are you sure you want to cancel this job run?"
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
                    headerText="Terminate Job Run?"
                    description="Are you sure you want to terminate this job run?"
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
              content={formatDateTime(
                jobRun?.startedAt ? timestampDate(jobRun.startedAt) : new Date()
              )}
            />
            <StatCard
              header="Completion Time"
              content={formatDateTime(
                jobRun?.completedAt
                  ? timestampDate(jobRun.completedAt)
                  : new Date()
              )}
            />
            <StatCard header="Duration" content={duration} />
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
            <JobRunActivityErrors
              jobRunId={id}
              jobId={jobRun?.jobId ?? ''}
              accountId={accountId}
            />
          </div>
          {!isSystemAppConfigDataLoading &&
            systemAppConfigData?.enableRunLogs && (
              <div>
                <JobRunLogs accountId={accountId} runId={id} />
              </div>
            )}
          <div className="space-y-4">
            <div className="flex flex-row items-center justify-end space-x-2">
              {isValidating && <Spinner />}
            </div>
            {eventsIsLoading ? (
              <SkeletonTable />
            ) : (
              <JobRunActivityTable
                jobRunEvents={eventData?.events}
                onViewSelectClicked={onViewSelectClicked}
                jobStatus={jobRun?.status}
              />
            )}
          </div>
          <ViewSelectDialog
            isDialogOpen={isViewSelectDialogOpen}
            setIsDialogOpen={setIsSelectDialogOpen}
            query={activeSelectQuery}
          />
        </div>
      )}
    </OverviewContainer>
  );
}

// There is way more here, but this is currently all we care about
interface RunContext {
  input: {
    pooled_sql_raw: {
      query: string;
    };
  };
}

function isValidRunContext(input: unknown): input is RunContext {
  const typedInput = input as Partial<RunContext>;
  return !!typedInput?.input?.pooled_sql_raw?.query;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function parseUint8ArrayToYaml(data: Uint8Array): any {
  try {
    const yamlString = new TextDecoder().decode(data);
    const result = yaml.parse(yamlString);
    return result;
  } catch (error) {
    console.error('Error parsing YAML:', error);
    return null;
  }
}

function buildBenthosRunCtxExternalId(schema: string, table: string): string {
  return `benthosconfig-${schema}.${table}.insert`;
}

interface SelectQuery {
  schema: string;
  table: string;
  select: string;
}

interface ViewSelectDialogProps {
  isDialogOpen: boolean;
  setIsDialogOpen(open: boolean): void;
  query: SelectQuery;
}

function ViewSelectDialog(props: ViewSelectDialogProps): ReactElement {
  const { isDialogOpen, setIsDialogOpen, query } = props;
  const { resolvedTheme } = useTheme();

  const formattedQuery = useMemo(() => {
    // todo: maybe update this to explicitly pass in the driver type so it formats it according to the correct connection
    return formatSql(query.select);
  }, [query.schema, query.table, query.select]);

  return (
    <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
      <DialogContent className="lg:max-w-4xl">
        <DialogHeader>
          <DialogTitle>
            SQL Select Query - {query.schema}.{query.table}
          </DialogTitle>
          <DialogDescription>
            This was the query used to query the source database during the job
            run
          </DialogDescription>
        </DialogHeader>
        <div>
          <Editor
            height="50vh"
            width="100%"
            language="sql"
            value={formattedQuery}
            theme={resolvedTheme === 'dark' ? 'vs-dark' : 'cobalt'}
            options={{
              minimap: { enabled: false },
              readOnly: true,
              wordWrap: 'on',
            }}
          />
        </div>
        <DialogFooter className="md:justify-between">
          <Button
            type="button"
            variant="secondary"
            onClick={() => setIsDialogOpen(false)}
          >
            Close
          </Button>
          <CopyButton
            buttonVariant="default"
            onCopiedText="Success!"
            onHoverText="Copy the SELECT Query"
            textToCopy={query.select}
            buttonText="Copy"
          />
        </DialogFooter>
      </DialogContent>
    </Dialog>
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

function getDuration(completedAt?: Date, startedAt?: Date): string {
  if (!startedAt || !completedAt) {
    return '';
  }

  const duration = intervalToDuration({ start: startedAt, end: completedAt });

  return formatDuration(duration, { format: ['minutes', 'seconds'] });
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
    return <div />;
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
