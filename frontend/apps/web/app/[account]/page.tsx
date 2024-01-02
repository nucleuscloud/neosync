'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Skeleton } from '@/components/ui/skeleton';
import Error from 'next/error';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';

export default function AccountPage(): ReactElement {
  const router = useRouter();
  const { account, isLoading } = useAccount();

  useEffect(() => {
    if (isLoading || !account?.name) {
      return;
    }
    router.push(`/${account.name}/jobs`);
  }, [isLoading, account?.name, account?.id]);

  if (isLoading) {
    return <Skeleton className="w-full h-full py-2" />;
  }

  if (!account) {
    return <Error statusCode={404} />;
  }

  return (
    <OverviewContainer
      Header={<PageHeader header={`Home - ${account.name}`} />}
      containerClassName="home-page"
    >
      <div className="flex flex-col gap-4">
        <SkeletonTable />
      </div>
    </OverviewContainer>
  );
}
