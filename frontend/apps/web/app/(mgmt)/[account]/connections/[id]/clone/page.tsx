'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useQuery } from '@connectrpc/connect-query';
import { Connection, ConnectionService } from '@neosync/sdk';
import Error from 'next/error';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useGetConnectionComponentDetails } from '../components/useGetConnectionComponentDetails';

export default function CloneConnectionPage({
  params,
}: PageProps): ReactElement {
  const id = params?.id ?? '';

  const { data: connection, isLoading } = useQuery(
    ConnectionService.method.getConnection,
    { id, excludeSensitive: false },
    { enabled: !!id }
  );

  if (!id) {
    return <Error statusCode={404} />;
  }

  if (isLoading) {
    return <SkeletonForm />;
  }

  if (!connection?.connection) {
    return <Error statusCode={404} />;
  }

  return <CloneForm connection={connection.connection} />;
}

interface CloneFormProps {
  connection: Connection;
}

function CloneForm(props: CloneFormProps): ReactElement {
  const { connection } = props;
  const router = useRouter();
  const { account } = useAccount();

  const connectionComponent = useGetConnectionComponentDetails({
    mode: 'clone',
    connection: {
      ...connection,
      name: `${connection.name}-copy`,
    },
    onSaved: (conn) => {
      router.push(`/${account?.name}/connections/${conn.id}`);
    },
  });

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
