'use client';
import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import PageHeader from '@/components/headers/PageHeader';
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

export default function JobIdLayout({ children, params }: LayoutProps) {
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
      router.push(`/jobs/${id}`);
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
        variant: 'destructive',
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
    {
      title: 'Subsets',
      href: `${basePath}/subsets`,
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
    <div className="p-5">
      <div className="flex flex-col gap-1">
        <h2 className="text-2xl font-bold tracking-tight">Job Overview</h2>
        <PageHeader
          header={data?.job?.name || ''}
          description={data?.job?.id || ''}
          extraHeading={
            <div className="flex flex-row space-x-4">
              <DeleteConfirmationDialog
                trigger={
                  <Button variant="destructive">
                    <ButtonText leftIcon={<TrashIcon />} text="Delete Job" />
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
      <div className="flex flex-col gap-6">
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
