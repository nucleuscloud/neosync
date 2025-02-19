'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import Error from 'next/error';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useGetConnectionComponentDetails } from '../components/useGetConnectionComponentDetails';
export default function CloneConnectionPage({
  params,
}: PageProps): ReactElement {
  const id = params?.id ?? '';

  const router = useRouter();
  const { account } = useAccount();

  const connectionComponent = useGetConnectionComponentDetails({
    mode: 'clone',
    connectionId: id,
    onSaved: (conn) => {
      router.push(`/${account?.name}/connections/${conn.id}`);
    },
  });

  if (!id) {
    return <Error statusCode={404} />;
  }

  return (
    <OverviewContainer
      Header={connectionComponent.header}
      containerClassName="px-32"
    >
      <div className="connection-details-container">
        <div className="flex flex-col gap-8">
          <div>{connectionComponent.body}</div>
        </div>
      </div>
    </OverviewContainer>
  );
}
