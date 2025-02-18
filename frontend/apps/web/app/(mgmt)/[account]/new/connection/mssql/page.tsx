'use client';
import SqlServerConnectionForm from '@/components/connections/forms/sql-server/SqlServerConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Connection } from '@neosync/sdk';
import { useRouter, useSearchParams } from 'next/navigation';
import { DiMysql } from 'react-icons/di';

export default function Mssql() {
  const onSuccess = useOnSuccess();
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Microsoft SQL Server"
          subHeadings="Configure a Microsoft SQL Server database as a connection"
          leftIcon={<DiMysql className="w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <SqlServerConnectionForm onSuccess={onSuccess} mode="create" />
    </OverviewContainer>
  );
}

function useOnSuccess(): (conn: Connection) => Promise<void> {
  const router = useRouter();
  const { account } = useAccount();
  // const posthog = usePostHog();
  const searchParams = useSearchParams();
  const returnTo = searchParams.get('returnTo');

  return async (conn: Connection): Promise<void> => {
    if (!account) {
      return;
    }
    if (returnTo) {
      router.push(returnTo);
    } else if (conn.id) {
      router.push(`/${account.name}/connections/${conn.id}`);
    }

    // try {
    //   // toast.success('Successfully created OpenAI Connection!');
    //   // posthog.capture('New Connection Created', { type: 'openai' });

    //   if (returnTo) {
    //     router.push(returnTo);
    //   } else if (conn.id) {
    //     router.push(`/${account.name}/connections/${conn.id}`);
    //   }
    // } catch (err) {
    //   console.error(err);
    //   toast.error('Unable to create OpenAI Connection', {
    //     description: getErrorMessage(err),
    //   });
    // }
  };
}
