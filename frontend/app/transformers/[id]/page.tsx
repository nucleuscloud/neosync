'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useToast } from '@/components/ui/use-toast';
import { useGetCustomTransformers } from '@/libs/hooks/useGetCustomTransformers';
import { getErrorMessage } from '@/util/util';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import { getTransformerComponentDetails } from './components/transformer-component';

export default function TransformerPage({ params }: PageProps) {
  const id = params?.id ?? '';
  const account = useAccount();
  const { data, isLoading, mutate } = useGetCustomTransformers(
    account?.id ?? ''
  );

  const { toast } = useToast();

  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }
  const tranformerComponent = getTransformerComponentDetails({
    transformer: data?.transformers.find((item) => item.id == id),
    onSaved: (resp) => {
      mutate();
      // new GetConnectionResponse({
      //   //udpate this to transformer
      //   connection: resp.connection,
      // })
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
