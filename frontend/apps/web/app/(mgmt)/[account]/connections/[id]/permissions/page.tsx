'use client';
import ResourceId from '@/components/ResourceId';
import Spinner from '@/components/Spinner';
import { SubNav } from '@/components/SubNav';
import OverviewContainer from '@/components/containers/OverviewContainer';
import LearnMoreLink from '@/components/labels/LearnMoreLink';
import PermissionsDataTable from '@/components/permissions/PermissionsDataTable';
import { TestConnectionResult } from '@/components/permissions/PermissionsDialog';
import {
  getPermissionColumns,
  PermissionConnectionType,
} from '@/components/permissions/columns';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import { useQuery } from '@connectrpc/connect-query';
import {
  ConnectionConfig,
  ConnectionConfigSchema,
  ConnectionRolePrivilege,
  ConnectionService,
} from '@neosync/sdk';
import { UpdateIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import Error from 'next/error';
import { useMemo, use } from 'react';
import { toast } from 'sonner';
import ViewActions from '../components/ViewActions';
import { useGetConnectionComponentDetails } from '../components/useGetConnectionComponentDetails';

function getPermissionColumnType(
  connConfig: ConnectionConfig
): PermissionConnectionType {
  switch (connConfig.config.case) {
    case 'mongoConfig':
      return 'mongodb';
    case 'mysqlConfig':
      return 'mysql';
    case 'pgConfig':
      return 'postgres';
    case 'dynamodbConfig':
      return 'dynamodb';
    default: // trash
      return 'postgres';
  }
}

export default function PermissionsPage(props: PageProps) {
  const params = use(props.params);
  const id = params?.id ?? '';
  const { account } = useAccount();
  const { data, isLoading } = useQuery(
    ConnectionService.method.getConnection,
    { id, excludeSensitive: true },
    { enabled: !!id }
  );

  const {
    data: connData,
    isLoading: isCheckConnLoading,
    isFetching,
    refetch: refetchCheckConnectionConfig,
  } = useQuery(
    ConnectionService.method.checkConnectionConfigById,
    {
      id,
    },
    { enabled: !!id }
  );

  const columns = useMemo(
    () =>
      getPermissionColumns(
        getPermissionColumnType(
          data?.connection?.connectionConfig ??
            create(ConnectionConfigSchema, {})
        )
      ),
    [isLoading]
  );

  const connectionComponent = useGetConnectionComponentDetails({
    mode: 'view',
    connection: data?.connection!,
    extraPageHeading: data?.connection && (
      <ViewActions connection={data.connection} />
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
  if (isLoading || isCheckConnLoading) {
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

  return (
    <OverviewContainer
      Header={connectionComponent.header}
      containerClassName="px-32"
    >
      <div className="connection-details-container">
        <div className="flex flex-col gap-8">
          <SubNav items={subnav} />
          <div>
            <PermissionsPageContainer
              data={connData?.privileges ?? []}
              isDbConnected={connData?.isConnected ?? false}
              connectionName={data?.connection?.name ?? ''}
              columns={columns}
              recheck={async () => {
                await refetchCheckConnectionConfig();
              }}
              isRechecking={isFetching}
            />
          </div>
        </div>
      </div>
    </OverviewContainer>
  );
}

interface PermissionsPageContainerProps {
  connectionName: string;
  data: ConnectionRolePrivilege[];
  isDbConnected: boolean;
  columns: ColumnDef<ConnectionRolePrivilege>[];
  recheck(): Promise<void>;
  isRechecking: boolean;
}

function PermissionsPageContainer(props: PermissionsPageContainerProps) {
  const {
    data,
    connectionName,
    isDbConnected,
    columns,
    recheck,
    isRechecking,
  } = props;

  async function handleMutate() {
    if (isRechecking) {
      return;
    }
    try {
      await recheck();
    } catch (error) {
      toast.error('Unable to get update permissions table', {
        description: getErrorMessage(error),
      });
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row justify-between items-center w-full">
        <div className="text-muted-foreground text-sm">
          Review the permissions that Neosync needs for your connection.{' '}
          <LearnMoreLink href="https://docs.neosync.dev/connections/postgres#permissions" />
        </div>
      </div>

      <PermissionsDataTable
        ConnectionAlert={
          <TestConnectionResult
            isConnected={isDbConnected}
            connectionName={connectionName}
            hasPrivileges={data.length > 0}
          />
        }
        TestConnectionButton={
          <Button type="button" variant="outline" onClick={handleMutate}>
            {isRechecking ? <Spinner /> : <UpdateIcon />}
          </Button>
        }
        data={data}
        columns={columns}
      />
    </div>
  );
}
