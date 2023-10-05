'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useToast } from '@/components/ui/use-toast';
import { useGetTransformers } from '@/libs/hooks/useGetTransformers';
import { getErrorMessage } from '@/util/util';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import { getTransformerComponentDetails } from './components/transformer-component';

export default function TransformerPage({ params }: PageProps) {
  const name = params?.name ?? '';
  const account = useAccount();
  const { data, isLoading } = useGetTransformers(account?.id ?? ''); //udpate with tranformesr

  const transformer = data?.transformers.find((item) => item.name == name);

  const { toast } = useToast();

  if (!name) {
    return <div>Can&apos;t find transformer ${name}</div>;
  }
  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }
  const tranformerComponent = getTransformerComponentDetails({
    transformer: transformer,
    // onSaved: (resp) => {
    //   mutate();
    //   new GetTransformersResponse({
    //     transformers: resp,
    //   });
    //   toast({
    //     title: 'Successfully updated transformer!',
    //     variant: 'default',
    //   });
    // },
    onSaveFailed: (err) =>
      toast({
        title: 'Unable to update transformer',
        description: getErrorMessage(err),
        variant: 'destructive',
      }),
    extraPageHeading: (
      <div>
        <RemoveTransformerButton transformerID={transformer?.id ?? ''} />
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
