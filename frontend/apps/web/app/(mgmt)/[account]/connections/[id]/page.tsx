'use client';
import ButtonText from '@/components/ButtonText';
import { CloneConnectionButton } from '@/components/CloneConnectionButton';
import ResourceId from '@/components/ResourceId';
import { SubNav } from '@/components/SubNav';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { useQuery } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import Error from 'next/error';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import RemoveConnectionButton from './components/RemoveConnectionButton';
import { useGetConnectionComponentDetails } from './components/connection-component';

export default function ConnectionPage({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { account } = useAccount();

  const { data, isLoading } = useQuery(
    ConnectionService.method.getConnection,
    { id: id, excludeSensitive: true },
    { enabled: !!id }
  );
  const router = useRouter();

  const connectionComponent = useGetConnectionComponentDetails({
    connection: data?.connection!,
    mode: 'view',
    extraPageHeading: (
      <div className="flex flex-row items-center gap-4">
        {data?.connection?.connectionConfig?.config.case &&
          data?.connection?.id && (
            <CloneConnectionButton id={data?.connection?.id ?? ''} />
          )}
        <RemoveConnectionButton connectionId={id} />
        <Button
          type="button"
          onClick={() => {
            router.push(`/${account?.name}/connections/${id}/edit`);
          }}
        >
          <ButtonText text="Edit" />
        </Button>
      </div>
    ),
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
