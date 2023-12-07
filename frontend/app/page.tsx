'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import Error from 'next/error';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';

export default function Home({ params }: PageProps): ReactElement {
  const router = useRouter();
  const { data, isLoading } = useGetUserAccounts();
  const accountName = params?.account ?? 'personal'; // if not present, may need to update url to include personal
  const account = data?.accounts.find((a) => a.name === accountName);

  useEffect(() => {
    if (isLoading || !account?.name) {
      return;
    }
    router.push(`/${account.name}/jobs`);
  }, [isLoading, accountName, account?.id]);

  if (isLoading) {
    return <Skeleton />;
  }

  if (!account) {
    return <Error statusCode={404} />;
  }

  return (
    <OverviewContainer
      Header={<PageHeader header={`Home - ${account.name}`} />}
      containerClassName="home-page"
    >
      <div className="flex flex-col gap-4">Hello</div>
    </OverviewContainer>
  );
}
