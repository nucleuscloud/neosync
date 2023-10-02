'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useToast } from '@/components/ui/use-toast';
import { useGetTransformers } from '@/libs/hooks/useGetTransformers';
import { GetConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { getErrorMessage } from '@/util/util';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import { getTransformerComponentDetails } from './components/transformer-component';

export default function TransformerPage({ params }: PageProps) {
  const id = params?.id ?? '';
  const { data, isLoading, mutate } = useGetTransformers(); //udpate with tranformesr

  const { toast } = useToast();
  if (!id) {
    return <div>Not Found</div>;
  }
  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }
  const tranformerComponent = getTransformerComponentDetails({
    transformer: data?.transformers[0]!,
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
        <RemoveTransformerButton transformerID={id} />
      </div>
    ),
  });
  return (
    <OverviewContainer Header={tranformerComponent.header}>
      <div className="connection-details-container">
        <div>
          <div className="flex flex-col">
            <div>{tranformerComponent.body}</div>
          </div>
        </div>
      </div>
    </OverviewContainer>
  );
}
