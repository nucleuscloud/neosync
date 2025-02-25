'use client';;
import { use } from "react";
import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import ResourceId from '@/components/ResourceId';
import { SubNav } from '@/components/SubNav';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { isJobSubsettable } from '@/components/jobs/subsets/utils';
import { useAccount } from '@/components/providers/account-provider';
import { LayoutProps } from '@/components/types';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getErrorMessage } from '@/util/util';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { Job, JobService, JobSourceOptions, JobStatus } from '@neosync/sdk';
import { LightningBoltIcon, TrashIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import JobIdSkeletonForm from './JobIdSkeletonForm';
import JobCloneButton from './components/JobCloneButton';
import JobPauseButton from './components/JobPauseButton';

export default function JobIdLayout(props: LayoutProps) {
  const params = use(props.params);

  const {
    children
  } = props;

  const id = params?.id ?? '';
  const router = useRouter();
  const { account } = useAccount();
  const { data, isLoading } = useQuery(
    JobService.method.getJob,
    { id },
    { enabled: !!id }
  );
  const { data: jobStatus, refetch: mutateJobStatus } = useQuery(
    JobService.method.getJobStatus,
    { jobId: id },
    { enabled: !!id }
  );
  const { refetch: mutateRecentRuns } = useQuery(
    JobService.method.getJobRecentRuns,
    { jobId: id },
    { enabled: !!id }
  );
  const { refetch: mutateJobRunsByJob } = useQuery(
    JobService.method.getJobRuns,
    { id: { case: 'jobId', value: id } },
    { enabled: !!id }
  );
  const { data: systemAppConfigData, isLoading: isSystemConfigLoading } =
    useGetSystemAppConfig();
  const { mutateAsync: removeJob } = useMutation(JobService.method.deleteJob);
  const { mutateAsync: triggerJobRun } = useMutation(
    JobService.method.createJobRun
  );

  async function onTriggerJobRun(): Promise<void> {
    try {
      await triggerJobRun({ jobId: id });
      toast.success('Job run triggered successfully!');
      setTimeout(() => {
        mutateRecentRuns();
        mutateJobRunsByJob();
      }, 4000); // delay briefly as there can sometimes be a trigger delay in temporal
    } catch (err) {
      console.error(err);
      toast.error('Uanble to trigger job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onDelete(): Promise<void> {
    if (!id) {
      return;
    }
    try {
      await removeJob({ id });
      toast.success('Job removed successfully!');
      router.push(`/${account?.name}/jobs`);
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove job', {
        description: getErrorMessage(err),
      });
    }
  }

  function onNewStatus(_newStatus: JobStatus): void {
    mutateJobStatus();
  }

  if (isLoading) {
    return (
      <div>
        <JobIdSkeletonForm />
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

  const sidebarNavItems = getSidebarNavItems(
    account?.name ?? '',
    data?.job,
    !isSystemConfigLoading && systemAppConfigData?.isMetricsServiceEnabled,
    !isSystemConfigLoading && systemAppConfigData?.isJobHooksEnabled
  );

  const badgeValue = getBadgeText(data.job.source?.options);

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
              leftBadgeValue={badgeValue}
              extraHeading={
                <div className="md:flex grid grid-cols-2 md:flex-row gap-4">
                  <JobCloneButton job={data.job} />
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

function getBadgeText(
  options?: JobSourceOptions
): 'Sync Job' | 'Generate Job' | 'AI Generate Job' {
  switch (options?.config.case) {
    case 'generate':
      return 'Generate Job';
    case 'aiGenerate':
      return 'AI Generate Job';
    default:
      return 'Sync Job';
  }
}

interface SidebarNav {
  title: string;
  href: string;
}
function getSidebarNavItems(
  accountName: string,
  job?: Job,
  isMetricsServiceEnabled?: boolean,
  isJobHooksEnabled?: boolean
): SidebarNav[] {
  if (!job) {
    return [{ title: 'Overview', href: `` }];
  }
  const basePath = `/${accountName}/jobs/${job.id}`;

  const nav = [
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
  ];

  if (isJobSubsettable(job)) {
    nav.push({
      title: 'Subsets',
      href: `${basePath}/subsets`,
    });
  }

  if (isMetricsServiceEnabled) {
    nav.push({
      title: 'Usage',
      href: `${basePath}/usage`,
    });
  }

  if (isJobHooksEnabled) {
    nav.push({ title: 'Hooks', href: `${basePath}/hooks` });
  }

  return nav;
}
