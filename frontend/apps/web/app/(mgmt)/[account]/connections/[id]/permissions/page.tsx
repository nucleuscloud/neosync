'use client';
'use client';
import { CloneConnectionButton } from '@/components/CloneConnectionButton';
import ResourceId from '@/components/ResourceId';
import Spinner from '@/components/Spinner';
import { SubNav } from '@/components/SubNav';
import OverviewContainer from '@/components/containers/OverviewContainer';
import LearnMoreTag from '@/components/labels/LearnMoreTag';
import PermissionsDataTable from '@/components/permissions/PermissionsDataTable';
import { TestConnectionResult } from '@/components/permissions/PermissionsDialog';
import { getPermissionColumns } from '@/components/permissions/columns';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { toast, useToast } from '@/components/ui/use-toast';
import { useGetConnection } from '@/libs/hooks/useGetConnection';
import { useTestProgressConnection } from '@/libs/hooks/useTestPostgresConnection';
import { getErrorMessage } from '@/util/util';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  CheckConnectionConfigResponse,
  ConnectionRolePrivilege,
  GetConnectionResponse,
  PostgresConnectionConfig,
} from '@neosync/sdk';
import { UpdateIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import Error from 'next/error';
import { useMemo } from 'react';
import { KeyedMutator } from 'swr';
import RemoveConnectionButton from '../components/RemoveConnectionButton';
import { getConnectionComponentDetails } from '../components/connection-component';

export default function PermissionsPage({ params }: PageProps) {
  const id = params?.id ?? '';
  const { account } = useAccount();
  const { data, isLoading, mutate } = useGetConnection(account?.id ?? '', id);

  const {
    data: validationRes,
    isLoading: isLoadingValidation,
    mutate: mutateValidation,
    isValidating: isValidating,
  } = useTestProgressConnection(
    account?.id ?? '',
    data?.connection?.connectionConfig?.config.case === 'pgConfig'
      ? data.connection.connectionConfig.config.value
      : new PostgresConnectionConfig({})
  );

  const { toast } = useToast();
  const columns = useMemo(() => getPermissionColumns(), []);

  if (!id) {
    return <Error statusCode={404} />;
  }
  if (isLoading || isLoadingValidation) {
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
      mutate(
        new GetConnectionResponse({
          connection: resp.connection,
        })
      );
      toast({
        title: 'Successfully updated connection!',
        variant: 'success',
      });
    },
    onSaveFailed: (err) =>
      toast({
        title: 'Unable to update connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      }),
    extraPageHeading: (
      <div className="flex flex-row items-center gap-4">
        {data?.connection?.connectionConfig?.config.case &&
          data?.connection?.id && (
            <CloneConnectionButton
              connectionType={
                data?.connection?.connectionConfig?.config.case ?? ''
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
              data={validationRes?.privileges ?? []}
              isDbConnected={validationRes?.isConnected ?? false}
              connectionName={data?.connection?.name ?? ''}
              columns={columns}
              mutateValidation={mutateValidation}
              isMutating={isValidating}
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
  columns: ColumnDef<PlainMessage<ConnectionRolePrivilege>>[];
  mutateValidation:
    | KeyedMutator<unknown>
    | KeyedMutator<CheckConnectionConfigResponse>;
  isMutating: boolean;
}

function PermissionsPageContainer(props: PermissionsPageContainerProps) {
  const {
    data,
    connectionName,
    isDbConnected,
    columns,
    mutateValidation,
    isMutating,
  } = props;

  const handleMutate = async () => {
    if (isMutating) {
      return;
    }
    try {
      await mutateValidation();
    } catch (error) {
      toast({
        title: 'Unable to update Permissions table!',
        variant: 'destructive',
      });
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row justify-between items-center w-full">
        <div className="text-muted-foreground text-sm">
          Review the permissions that Neosync needs for your connection.{' '}
          <LearnMoreTag href="https://docs.neosync.dev/connections/postgres#permissions" />
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
            {isMutating ? <Spinner /> : <UpdateIcon />}
          </Button>
        }
        data={data}
        columns={columns}
      />
    </div>
  );
}
