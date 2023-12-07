'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import Error from 'next/error';
import { ReactElement } from 'react';

export default function AccountPage({ params }: PageProps): ReactElement {
  const { data, isLoading } = useGetUserAccounts();
  const accountName = params?.account ?? 'personal'; // if not present, may need to update url to include personal

  if (isLoading) {
    return <Skeleton />;
  }

  const account = data?.accounts.find((a) => a.name === accountName);
  if (!account) {
    return <Error statusCode={404} />;
  }
  // const authEnabled = useGetAuthEnabled();

  // const router = useRouter();

  // useEffect(() => {
  //   if (!authEnabled) {
  //     router.push('/settings/temporal');
  //   }
  // });
  return (
    <OverviewContainer
      Header={<PageHeader header={`Home - ${account.name}`} />}
      containerClassName="home-page"
    >
      <div className="flex flex-col gap-4">Hello</div>
    </OverviewContainer>
  );
}
