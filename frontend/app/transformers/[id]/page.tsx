'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useToast } from '@/components/ui/use-toast';
import { useGetCustomTransformersById } from '@/libs/hooks/useGetCustomTransformerById';
import { GetCustomTransformersResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import { getTransformerComponentDetails } from './components/transformer-component';

export default function NewCustomTransformerPage({ params }: PageProps) {
  const id = params?.id ?? '';

  const { data, isLoading, mutate } = useGetCustomTransformersById(id);

  const { toast } = useToast();

  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }

  const tranformerComponent = getTransformerComponentDetails({
    CustomTransformer: data?.transformer,
    onSaved: (resp) => {
      mutate();
      new GetCustomTransformersResponse({
        //udpate this to transformer
        // transformers: resp.connection,
      });
      console.log('resp', resp);
      toast({
        title: 'Successfully updated transformer!',
        variant: 'default',
      });
    },
    onSaveFailed: (err) =>
      toast({
        title: 'Unable to update transformer',
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
      <div className="transformer-details-container">
        <div>
          <div className="flex flex-col">
            <div>{tranformerComponent.body}</div>
          </div>
        </div>
      </div>
    </OverviewContainer>
  );
}
