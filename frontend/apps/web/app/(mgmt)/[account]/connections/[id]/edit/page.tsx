'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import ResourceId from '@/components/ResourceId';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { SubNav } from '@/components/SubNav';
import { PageProps } from '@/components/types';
import { useQuery } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import Error from 'next/error';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useGetConnectionComponentDetails } from '../components/useGetConnectionComponentDetails';

export default function EditConnectionPage({ params }: PageProps) {
  const id = params?.id ?? '';
  const { account } = useAccount();

  // todo: add check to ensure user has permission to edit connection
  // if not, do not render component or load query

  const {
    data,
    isLoading,
    refetch: mutateGetConnection,
  } = useQuery(
    ConnectionService.method.getConnection,
    { id: id, excludeSensitive: false },
    { enabled: !!id }
  );

  const router = useRouter();

  const connectionComponent = useGetConnectionComponentDetails({
    mode: 'edit',
    connection: data?.connection!,
    onSaved: (updatedConnection) => {
      toast.success('Successfully updated connection!');
      // maybe use the query cache here for faster round trip and navigation
      mutateGetConnection().finally(() => {
        router.push(`/${account?.name}/connections/${updatedConnection.id}`);
      });
    },
    subHeading: (
      <ResourceId
        labelText={data?.connection?.id ?? ''}
        copyText={data?.connection?.id ?? ''}
        onHoverText="Copy the connection id"
      />
    ),
  });

  if (!id) {
    return <Error statusCode={404} />;
  }

  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }
  if (!isLoading && !data?.connection) {
    return <Error statusCode={404} />;
  }
  const basePath = `/${account?.name}/connections/${data?.connection?.id}`;

  const subnav = [
    {
      title: 'Configuration',
      href: `${basePath}`,
    },
    {
      title: 'Permissions',
      href: `${basePath}/permissions`,
    },
  ];

  const showSubNav =
    data?.connection?.connectionConfig?.config.case === 'pgConfig' ||
    data?.connection?.connectionConfig?.config.case === 'mysqlConfig' ||
    data?.connection?.connectionConfig?.config.case === 'dynamodbConfig' ||
    data?.connection?.connectionConfig?.config.case === 'mongoConfig';

  return (
    <OverviewContainer
      Header={connectionComponent.header}
      containerClassName="px-32"
    >
      <div className="connection-details-container">
        <div className="flex flex-col gap-8">
          {showSubNav && <SubNav items={subnav} buttonClassName="" />}
          <div>{connectionComponent.body}</div>
        </div>
      </div>
    </OverviewContainer>
  );
}
