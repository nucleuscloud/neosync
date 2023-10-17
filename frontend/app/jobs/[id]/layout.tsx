'use client';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import SubPageHeader from '@/components/headers/SubPageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { LayoutProps } from '@/components/types';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button, buttonVariants } from '@/components/ui/button';
import { toast } from '@/components/ui/use-toast';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { TrashIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';

export default function SettingsLayout({ children, params }: LayoutProps) {
  const id = params?.id ?? '';
  const basePath = `/jobs/${params?.id}`;
  const { data, isLoading } = useGetJob(id);
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

  async function onDelete(): Promise<void> {
    if (!id) {
      return;
    }
    try {
      await removeJob(id);
      toast({
        title: 'Job removed successfully!',
      });
      router.push('/jobs');
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove job',
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
            <div className="flex flex-row space-x-4">
              <DeleteConfirmationDialog
                trigger={
                  <Button variant="destructive" size="icon">
                    <TrashIcon />
                  </Button>
                }
                headerText="Are you sure you want to delete this job?"
                description="Deleting this job will also delete all job runs."
                onConfirm={async () => onDelete()}
              />
              <Button onClick={() => onTriggerJobRun()}>Trigger Run</Button>
            </div>
          }
        />
      </div>
      <div className="flex flex-col space-y-8">
        <SubNav items={sidebarNavItems} />
        <div>{children}</div>
      </div>
    </div>
  );
}

async function removeJob(jobId: string): Promise<void> {
  const res = await fetch(`/api/jobs/${jobId}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
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
        'flex flex-col lg:flex-row space-y-2 lg:space-y-0 lg:space-x-2',
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
                : 'hover:bg-transparent hover:underline',
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
