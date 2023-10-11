'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useGetTransformers } from '@/libs/hooks/useGetTransformers';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import { getTransformerComponentDetails } from './components/transformer-component';

export default function TransformerPage({ params }: PageProps) {
  const title = params?.name ?? '';

  const { data, isLoading } = useGetTransformers();

  const transformer = data?.transformers.find((item) => item.title == title);

  // const { toast } = useToast();

  if (!title) {
    return <div>Can&apos;t find transformer ${title}</div>;
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
    // onSaveFailed: (err) =>
    //   toast({
    //     title: 'Unable to update transformer',
    //     description: getErrorMessage(err),
    //     variant: 'destructive',
    //   }),
    extraPageHeading: (
      <div>
        <RemoveTransformerButton transformerID={transformer?.title ?? ''} />
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
