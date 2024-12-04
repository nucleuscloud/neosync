import SubPageHeader from '@/components/headers/SubPageHeader';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery } from '@connectrpc/connect-query';
import { getJobHooks } from '@neosync/sdk/connectquery';
import { ReactElement } from 'react';
import HookCard from './HookCard';

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

  if (isGetJobHooksLoading) {
    return (
      <div>
        <Skeleton />
      </div>
    );
  }

  const jobHooks = getJobHooksResp?.hooks ?? [];

  return (
    <div className="job-hooks-card-container flex flex-col gap-3">
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
        extraHeading={<Button type="button">New Hook</Button>}
      />

      <div className="flex flex-col gap-5">
        {jobHooks.map((hook) => {
          return (
            <HookCard
              key={hook.id}
              hook={hook}
              onDeleted={refetch}
              onEdited={refetch}
            />
          );
        })}
      </div>
    </div>
  );
}
