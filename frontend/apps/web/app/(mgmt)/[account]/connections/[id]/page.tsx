'use client';
import { CloneConnectionButton } from '@/components/CloneConnectionButton';
import ResourceId from '@/components/ResourceId';
import { SubNav } from '@/components/SubNav';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { getErrorMessage } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import { createConnectQueryKey, useQuery } from '@connectrpc/connect-query';
import {
  ConnectionConfigSchema,
  ConnectionService,
  GetConnectionResponseSchema,
} from '@neosync/sdk';
import { useQueryClient } from '@tanstack/react-query';
import Error from 'next/error';
import { toast } from 'sonner';
import RemoveConnectionButton from './components/RemoveConnectionButton';
import { getConnectionComponentDetails } from './components/connection-component';

export default function ConnectionPage({ params }: PageProps) {
  const id = params?.id ?? '';
  const { account } = useAccount();

  const { data, isLoading } = useQuery(
    ConnectionService.method.getConnection,
    { id: id },
    { enabled: !!id }
  );
  const queryclient = useQueryClient();
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
  const connectionComponent = getConnectionComponentDetails({
    connection: data?.connection!,
    onSaved: (resp) => {
      const key = createConnectQueryKey({
        schema: ConnectionService.method.getConnection,
        input: { id },
        cardinality: undefined,
      });
      queryclient.setQueryData(
        key,
        create(GetConnectionResponseSchema, { connection: resp.connection })
      );
      toast.success('Successfully updated connection!');
    },
    onSaveFailed: (err) =>
      toast.error('Unable to update connection', {
        description: getErrorMessage(err),
      }),
    extraPageHeading: (
      <div className="flex flex-row items-center gap-4">
        {data?.connection?.connectionConfig?.config.case &&
          data?.connection?.id && (
            <CloneConnectionButton
              connectionConfig={
                data?.connection?.connectionConfig ??
                create(ConnectionConfigSchema, {})
              }
              id={data?.connection?.id ?? ''}
            />
          )}
        <RemoveConnectionButton connectionId={id} />
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
