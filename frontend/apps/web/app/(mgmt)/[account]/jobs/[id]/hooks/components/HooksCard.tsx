import SubPageHeader from '@/components/headers/SubPageHeader';
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
    <div className="job-hooks-card-container">
      <SubPageHeader
        header="Job Hooks"
        description="Manage a job's hooks"
        extraHeading={<Button type="button">New Hook</Button>}
      />

      <div className="flex flex-col gap-5">
        {jobHooks.map((hook) => {
          return <HookCard key={hook.id} hook={hook} />;
        })}
      </div>
    </div>
  );
}
