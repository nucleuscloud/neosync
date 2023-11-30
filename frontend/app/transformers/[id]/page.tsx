'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useGetCustomTransformersById } from '@/libs/hooks/useGetCustomTransformerById';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import UpdateTransformerForm from './components/UpdateTransformerForm';

export default function NewCustomTransformerPage({ params }: PageProps) {
  const id = params?.id ?? '';

  const { data, isLoading } = useGetCustomTransformersById(id);

  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header={data?.transformer?.name ?? 'Custom Transformer'}
          leftBadgeValue={data?.transformer?.type}
          extraHeading={
            <RemoveTransformerButton
              transformerID={data?.transformer?.id ?? ''}
            />
          }
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <div className="transformer-details-container">
        <div>
          <div className="flex flex-col">
            <div>
              <UpdateTransformerForm currentTransformer={data?.transformer} />
            </div>
          </div>
        </div>
      </div>
    </OverviewContainer>
  );
}
