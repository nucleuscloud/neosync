import EmptyState from '@/components/EmptyState';
import SubPageHeader from '@/components/headers/SubPageHeader';
import Spinner from '@/components/Spinner';
import { Skeleton } from '@/components/ui/skeleton';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { useQuery } from '@connectrpc/connect-query';
import { AccountHookService } from '@neosync/sdk';
import { ReactElement, ReactNode, useMemo } from 'react';
import { MdWebhook } from 'react-icons/md';
import HookCard from './HookCard';
import NewHookButton from './NewHookButton';

interface Props {
  accountId: string;
}

export default function HooksCard(props: Props): ReactElement<any> {
  const { accountId } = props;
  const {
    data: getAccountHooksResp,
    isLoading: isGetAccountHooksLoading,
    isFetching: isGetAccountHooksFetching,
    refetch,
  } = useQuery(
    AccountHookService.method.getAccountHooks,
    { accountId: accountId },
    { enabled: !!accountId }
  );

  const accountHooks = useMemo(() => {
    return [...(getAccountHooksResp?.hooks ?? [])].sort((a, b) => {
      const timeA = a.createdAt ? timestampDate(a.createdAt).getTime() : 0;
      const timeB = b.createdAt ? timestampDate(b.createdAt).getTime() : 0;
      return timeA - timeB;
    });
  }, [getAccountHooksResp?.hooks, isGetAccountHooksFetching]);

  if (isGetAccountHooksLoading) {
    return (
      <div>
        <Skeleton />
      </div>
    );
  }

  return (
    <div className="job-hooks-card-container flex flex-col gap-5">
      <SubPageHeader
        header="Account Hooks"
        rightHeaderIcon={
          isGetAccountHooksFetching ? (
            <div>
              <Spinner className="h-4 w-4" />
            </div>
          ) : null
        }
        description="Manage hooks that execute when events occur"
        extraHeading={
          <NewHookButton accountId={accountId} onCreated={refetch} />
        }
      />

      <div className="flex flex-col gap-5">
        {accountHooks.length === 0 && (
          <NoAccountHooks
            button={<NewHookButton accountId={accountId} onCreated={refetch} />}
          />
        )}
        {accountHooks.map((hook) => {
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

interface NoAccountHooksProps {
  button: ReactNode;
}

function NoAccountHooks(props: NoAccountHooksProps): ReactElement<any> {
  const { button } = props;
  return (
    <EmptyState
      title="No Hooks yet"
      description="Hooks are invoked when specified events occur in your account to notify you or trigger further actions"
      icon={<MdWebhook className="w-8 h-8 text-primary" />}
      extra={button}
    />
  );
}
