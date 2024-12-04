import SubPageHeader from '@/components/headers/SubPageHeader';
import Spinner from '@/components/Spinner';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery } from '@connectrpc/connect-query';
import { Job } from '@neosync/sdk';
import { getJob, getJobHooks } from '@neosync/sdk/connectquery';
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

  const jobHooks = useMemo(() => {
    return [...(getJobHooksResp?.hooks ?? [])].sort((a, b) => {
      const timeA = a.createdAt ? a.createdAt.toDate().getTime() : 0;
      const timeB = b.createdAt ? b.createdAt.toDate().getTime() : 0;
      return timeA - timeB;
    });
  }, [getJobHooksResp?.hooks, isGetJobHooksFetching]);

  const jobConnectionIds = useMemo(() => {
    return getJobConnectionIds(getJobResp?.job ?? new Job());
  }, [getJobResp?.job, isGetJobFetching]);

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
          isGetJobHooksFetching ? (
            <div>
              <Spinner className="h-4 w-4" />
            </div>
          ) : null
        }
        description="Manage a job's hooks"
        extraHeading={
          <NewHookButton
            jobId={jobId}
            jobConnectionIds={jobConnectionIds}
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
              jobConnectionIds={jobConnectionIds}
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
