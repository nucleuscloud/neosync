'use client';
import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import ResourceId from '@/components/ResourceId';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { LayoutProps } from '@/components/types';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button, buttonVariants } from '@/components/ui/button';
import { toast } from '@/components/ui/use-toast';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { useGetJobRecentRuns } from '@/libs/hooks/useGetJobRecentRuns';
import { useGetJobRunsByJob } from '@/libs/hooks/useGetJobRunsByJob';
import { useGetJobStatus } from '@/libs/hooks/useGetJobStatus';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { GetJobStatusResponse, Job, JobStatus } from '@neosync/sdk';
import { LightningBoltIcon, TrashIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import JobPauseButton from './components/JobPauseButton';
import { isDataGenJob } from './util';

export default function JobIdLayout({ children, params }: LayoutProps) {
  const id = params?.id ?? '';
  const router = useRouter();
  const { account } = useAccount();
  const { data, isLoading } = useGetJob(account?.id ?? '', id);
  const { data: jobStatus, mutate: mutateJobStatus } = useGetJobStatus(
    account?.id ?? '',
    id
  );
  const { mutate: mutateRecentRuns } = useGetJobRecentRuns(
    account?.id ?? '',
    id
  );
  const { mutate: mutateJobRunsByJob } = useGetJobRunsByJob(
    account?.id ?? '',
    id
  );

  const { data: systemAppConfigData, isLoading: isSystemConfigLoading } =
    useGetSystemAppConfig();

  async function onTriggerJobRun(): Promise<void> {
    try {
      await triggerJobRun(account?.id ?? '', id);
      toast({
        title: 'Job run triggered successfully!',
        variant: 'success',
      });
      setTimeout(() => {
        mutateRecentRuns();
        mutateJobRunsByJob();
      }, 3000); // delay briefly as there can sometimes be a trigger delay in temporal
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to trigger job run',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function onDelete(): Promise<void> {
    if (!id) {
      return;
    }
    try {
      await removeJob(account?.id ?? '', id);
      toast({
        title: 'Job removed successfully!',
        variant: 'success',
      });
      router.push(`/${account?.name}/jobs`);
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove job',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  function onNewStatus(newStatus: JobStatus): void {
    mutateJobStatus(new GetJobStatusResponse({ status: newStatus }));
  }

  if (isLoading) {
    return (
      <div>
        <SkeletonForm />
      </div>
    );
  }

  if (!data?.job) {
    return (
      <div className="mt-8">
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve job`}</AlertTitle>
        </Alert>
      </div>
    );
  }

  let sidebarNavItems = getSidebarNavItems(account?.name ?? '', data?.job);
  sidebarNavItems =
    isSystemConfigLoading || !systemAppConfigData?.isMetricsServiceEnabled
      ? sidebarNavItems.filter((item) => !item.href.endsWith('/usage'))
      : sidebarNavItems;

  return (
    <div>
      <OverviewContainer
        Header={
          <div>
            <PageHeader
              pageHeaderContainerClassName="gap-2"
              header={data?.job?.name || ''}
              subHeadings={
                <ResourceId
                  labelText={data?.job?.id ?? ''}
                  copyText={data?.job?.id ?? ''}
                  onHoverText="Copy the Job ID"
                />
              }
              leftBadgeValue={
                data.job.source?.options?.config.case === 'generate'
                  ? 'Generate Job'
                  : 'Sync Job'
              }
              extraHeading={
                <div className="flex flex-row space-x-4">
                  <DeleteConfirmationDialog
                    trigger={
                      <Button variant="destructive">
                        <ButtonText
                          leftIcon={<TrashIcon />}
                          text="Delete Job"
                        />
                      </Button>
                    }
                    headerText="Are you sure you want to delete this job?"
                    description="Deleting this job will also delete all job runs."
                    onConfirm={async () => onDelete()}
                  />
                  <JobPauseButton
                    jobId={id}
                    status={jobStatus?.status}
                    onNewStatus={onNewStatus}
                  />
                  <Button onClick={() => onTriggerJobRun()}>
                    <ButtonText
                      leftIcon={<LightningBoltIcon />}
                      text="Trigger Run"
                    />
                  </Button>
                </div>
              }
            />
          </div>
        }
      >
        <div className="flex flex-col gap-12">
          <SubNav items={sidebarNavItems} />
          <div>{children}</div>
        </div>
      </OverviewContainer>
    </div>
  );
}

interface SidebarNav {
  title: string;
  href: string;
}
function getSidebarNavItems(accountName: string, job?: Job): SidebarNav[] {
  if (!job) {
    return [{ title: 'Overview', href: `` }];
  }
  const basePath = `/${accountName}/jobs/${job.id}`;

  if (isDataGenJob(job)) {
    return [
      {
        title: 'Overview',
        href: `${basePath}`,
      },
      {
        title: 'Source',
        href: `${basePath}/source`,
      },
      {
        title: 'Destinations',
        href: `${basePath}/destinations`,
      },
      {
        title: 'Usage',
        href: `${basePath}/usage`,
      },
    ];
  }

  return [
    {
      title: 'Overview',
      href: `${basePath}`,
    },
    {
      title: 'Source',
      href: `${basePath}/source`,
    },
    {
      title: 'Destinations',
      href: `${basePath}/destinations`,
    },
    {
      title: 'Subsets',
      href: `${basePath}/subsets`,
    },
    {
      title: 'Usage',
      href: `${basePath}/usage`,
    },
  ];
}

async function removeJob(accountId: string, jobId: string): Promise<void> {
  const res = await fetch(`/api/accounts/${accountId}/jobs/${jobId}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

async function triggerJobRun(accountId: string, jobId: string): Promise<void> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/create-run`,
    {
      method: 'POST',
      body: JSON.stringify({ jobId }),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

interface SubNavProps extends React.HTMLAttributes<HTMLElement> {
  items: {
    href: string;
    title: string;
  }[];
  buttonClassName?: string;
}

function SubNav({ className, items, buttonClassName, ...props }: SubNavProps) {
  const pathname = usePathname();
  return (
    <nav
      className={cn(
        'flex flex-col lg:flex-row gap-2 lg:gap-y-0 lg:gap-x-2',
        className
      )}
      {...props}
    >
      {items.map((item) => {
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              buttonVariants({ variant: 'outline' }),
              pathname === item.href
                ? 'bg-muted hover:bg-muted'
                : 'hover:bg-gray-100 dark:hover:bg-gray-800',
              'justify-start',
              buttonClassName,
              'px-6'
            )}
          >
            {item.title}
          </Link>
        );
      })}
    </nav>
  );
}
