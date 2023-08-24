'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { PageProps } from '@/components/types';
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnection } from '@/libs/hooks/useGetConnection';
import { GetConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { getErrorMessage } from '@/util/util';
import RemoveConnectionButton from './components/RemoveConnectionButton';
import { getConnectionComponentDetails } from './components/connection-component';

export default function ConnectionPage({ params }: PageProps) {
  const id = params?.id ?? '';
  const { data, isLoading, mutate } = useGetConnection(id);
  const { toast } = useToast();
  if (!id) {
    return <div>Not Found</div>;
  }
  if (isLoading) {
    return <Skeleton className="w-[100px] h-[20px] rounded-full" />;
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
        variant: 'default',
      });
    },
    onSaveFailed: (err) =>
      toast({
        title: 'Unable to update connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      }),
    extraPageHeading: (
      <div>
        <RemoveConnectionButton connectionId={id} />
      </div>
    ),
  });
  return (
    <OverviewContainer Header={connectionComponent.header}>
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
