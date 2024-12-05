import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import Spinner from '@/components/Spinner';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery } from '@connectrpc/connect-query';
import { Job } from '@neosync/sdk';
import { getConnections, getJob, getJobHooks } from '@neosync/sdk/connectquery';
import { ReactElement, useMemo } from 'react';
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
  } = useQuery(getJobHooks, { jobId: jobId }, { enabled: !!jobId });

  const {
    data: getJobResp,
    isLoading: isGetJobLoading,
    isFetching: isGetJobFetching,
  } = useQuery(getJob, { id: jobId }, { enabled: !!jobId });

  const { account } = useAccount();
  const { data: getConnectionsResp, isFetching: isConnectionsFetching } =
    useQuery(
      getConnections,
      { accountId: account?.id },
      { enabled: !!account?.id }
    );

  const jobHooks = useMemo(() => {
    return [...(getJobHooksResp?.hooks ?? [])].sort((a, b) => {
      const timeA = a.createdAt ? a.createdAt.toDate().getTime() : 0;
      const timeB = b.createdAt ? b.createdAt.toDate().getTime() : 0;
      return timeA - timeB;
    });
  }, [getJobHooksResp?.hooks, isGetJobHooksFetching]);

  const jobConnectionIds = useMemo(() => {
    return new Set(getJobConnectionIds(getJobResp?.job ?? new Job()));
  }, [getJobResp?.job, isGetJobFetching]);

  const jobConnections = useMemo(() => {
    return (
      getConnectionsResp?.connections.filter((conn) =>
        jobConnectionIds.has(conn.id)
      ) ?? []
    );
  }, [isConnectionsFetching, jobConnectionIds]);

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
        {jobHooks.map((hook) => {
          return (
            <HookCard
              key={hook.id}
              hook={hook}
              onDeleted={refetch}
              onEdited={refetch}
              jobConnections={jobConnections}
            />
          );
        })}
      </div>
    </div>
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
