import EmptyState from '@/components/EmptyState';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import Spinner from '@/components/Spinner';
import { Skeleton } from '@/components/ui/skeleton';
import { create } from '@bufbuild/protobuf';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { useQuery } from '@connectrpc/connect-query';
import { ConnectionService, Job, JobSchema, JobService } from '@neosync/sdk';
import { ReactElement, ReactNode, useMemo } from 'react';
import { MdWebhook } from 'react-icons/md';
import { getConnectionIdFromSource } from '../../source/components/util';
import HookCard from './HookCard';
import NewHookButton from './NewHookButton';

interface Props {
  jobId: string;
}

export default function HooksCard(props: Props): ReactElement {
  const { jobId } = props;
  const {
    data: getJobHooksResp,
    isLoading: isGetJobHooksLoading,
    isFetching: isGetJobHooksFetching,
    refetch,
  } = useQuery(
    JobService.method.getJobHooks,
    { jobId: jobId },
    { enabled: !!jobId }
  );

  const {
    data: getJobResp,
    isLoading: isGetJobLoading,
    isFetching: isGetJobFetching,
  } = useQuery(JobService.method.getJob, { id: jobId }, { enabled: !!jobId });

  const { account } = useAccount();
  const { data: getConnectionsResp, isFetching: isConnectionsFetching } =
    useQuery(
      ConnectionService.method.getConnections,
      { accountId: account?.id },
      { enabled: !!account?.id }
    );

  const jobHooks = useMemo(() => {
    return [...(getJobHooksResp?.hooks ?? [])].sort((a, b) => {
      const timeA = a.createdAt ? timestampDate(a.createdAt).getTime() : 0;
      const timeB = b.createdAt ? timestampDate(b.createdAt).getTime() : 0;
      return timeA - timeB;
    });
  }, [getJobHooksResp?.hooks, isGetJobHooksFetching]);

  const jobConnectionIds = useMemo(() => {
    return new Set(getJobConnectionIds(getJobResp?.job ?? create(JobSchema)));
  }, [getJobResp?.job, isGetJobFetching]);

  const jobConnections = useMemo(() => {
    return (
      getConnectionsResp?.connections.filter((conn) =>
        jobConnectionIds.has(conn.id)
      ) ?? []
    );
  }, [isConnectionsFetching, jobConnectionIds]);

  const connectionMap = useMemo(
    () => new Map(jobConnections.map((conn) => [conn.id, conn])),
    [jobConnections]
  );

  if (isGetJobHooksLoading || isGetJobLoading) {
    return (
      <div>
        <Skeleton />
      </div>
    );
  }

  return (
    <div className="job-hooks-card-container flex flex-col gap-5">
      <SubPageHeader
        header="Job Hooks"
        rightHeaderIcon={
          isGetJobHooksFetching || isGetJobFetching ? (
            <div>
              <Spinner className="h-4 w-4" />
            </div>
          ) : null
        }
        description="Manage hooks that execute at specific points in a job run's lifecycle"
        extraHeading={
          <NewHookButton
            jobId={jobId}
            jobConnections={jobConnections}
            onCreated={refetch}
          />
        }
      />

      <div className="flex flex-col gap-5">
        {jobHooks.length === 0 && (
          <NoJobHooks
            button={
              <NewHookButton
                jobId={jobId}
                jobConnections={jobConnections}
                onCreated={refetch}
              />
            }
          />
        )}
        {jobHooks.map((hook) => {
          return (
            <HookCard
              key={hook.id}
              hook={hook}
              onDeleted={refetch}
              onEdited={refetch}
              jobConnections={jobConnections}
              jobConnectionsMap={connectionMap}
            />
          );
        })}
      </div>
    </div>
  );
}

interface NoJobHooksProps {
  button: ReactNode;
}

function NoJobHooks(props: NoJobHooksProps): ReactElement {
  const { button } = props;
  return (
    <EmptyState
      title="No Hooks yet"
      description="Hooks are events that are invoked during a job run to allow for further customization of a sync"
      icon={<MdWebhook className="w-8 h-8 text-primary" />}
      extra={button}
    />
  );
}

function getJobConnectionIds(job: Job): string[] {
  const output: string[] = [];

  const sourceId = getConnectionIdFromSource(job.source);
  if (sourceId) {
    output.push(sourceId);
  }
  return output.concat(job.destinations.map((dest) => dest.connectionId));
}
