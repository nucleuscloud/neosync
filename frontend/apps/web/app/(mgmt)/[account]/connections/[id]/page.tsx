'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnection } from '@/libs/hooks/useGetConnection';
import { getErrorMessage } from '@/util/util';
import { GetConnectionResponse } from '@neosync/sdk';
import Error from 'next/error';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { GrClone } from 'react-icons/gr';
import RemoveConnectionButton from './components/RemoveConnectionButton';
import { getConnectionComponentDetails } from './components/connection-component';

export default function ConnectionPage({ params }: PageProps) {
  const id = params?.id ?? '';
  const { account } = useAccount();
  const { data, isLoading, mutate } = useGetConnection(account?.id ?? '', id);
  const { toast } = useToast();
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
        <CloneConnectionButton
          connectionType={data?.connection?.connectionConfig?.config.case ?? ''}
          id={data?.connection?.id ?? ''}
        />
        <RemoveConnectionButton connectionId={id} />
      </div>
    ),
  });
  return (
    <OverviewContainer
      Header={connectionComponent.header}
      containerClassName="px-32"
    >
      <div className="connection-details-container">
        <div>
          <div className="flex flex-col">
            <div>{connectionComponent.body}</div>
          </div>
        </div>
      </div>
    </OverviewContainer>
  );
}

interface CloneConnectionProps {
  connectionType: string;
  id: string;
}

function CloneConnectionButton(props: CloneConnectionProps): ReactElement {
  const { connectionType, id } = props;
  const { account } = useAccount();

  let connType = '';

  switch (connectionType) {
    case 'pgConfig':
      connType = 'postgres';
    case 'mysqlConfig':
      connType = 'mysql';
    case 'awsS3Config':
      connType = 'aws-s3';
    default:
      connType = 'postgres';
  }

  return (
    <NextLink
      href={`/${account?.name}/new/connection/${connType}?sourceId=${id}`}
    >
      <Button>
        <ButtonText
          text="Clone Connection"
          leftIcon={<GrClone className="mr-1" />}
        />
      </Button>
    </NextLink>
  );
}
