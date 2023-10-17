'use client';
import { SidebarNav } from '@/components/SideBarNav';
import SubPageHeader from '@/components/headers/SubPageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { toast } from '@/components/ui/use-toast';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { getErrorMessage } from '@/util/util';
import { useRouter } from 'next/navigation';

export default function SettingsLayout({ children, params }: PageProps) {
  const id = params?.id ?? '';
  const basePath = `/jobs/${params?.id}`;
  const { data, isLoading, mutate } = useGetJob(id);
  const router = useRouter();

  async function onTriggerJobRun(): Promise<void> {
    try {
      await triggerJobRun(id);
      toast({
        title: 'Job run triggered successfully!',
      });
      router.push(`/runs`);
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to trigger job run',
        description: getErrorMessage(err),
      });
    }
  }

  const sidebarNavItems = [
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

  if (isLoading) {
    return (
      <div>
        <SkeletonForm />
      </div>
    );
  }

  if (!data?.job) {
    return (
      <div className="mt-10">
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve job`}</AlertTitle>
        </Alert>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-10 pb-16 md:block">
      <div className="space-y-1">
        <h2 className="text-2xl font-bold tracking-tight">Job Overview</h2>
        <SubPageHeader
          header={data?.job?.name || ''}
          description={data?.job?.id || ''}
          extraHeading={
            <Button onClick={() => onTriggerJobRun()}>Trigger Run</Button>
          }
        />
      </div>
      <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-12 lg:space-y-0">
        <aside className="-mx-4 lg:w-[200px]">
          <SidebarNav items={sidebarNavItems} />
        </aside>
        <div className="flex-1 lg:max-w-8xl">{children}</div>
      </div>
    </div>
  );
}

async function triggerJobRun(jobId: string): Promise<void> {
  const res = await fetch(`/api/jobs/${jobId}/create-run`, {
    method: 'POST',
    body: JSON.stringify({ jobId }),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
