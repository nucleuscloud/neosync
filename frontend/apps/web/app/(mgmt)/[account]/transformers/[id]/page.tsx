'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { useGetUserDefinedTransformersById } from '@/libs/hooks/useGetUserDefinedTransformerById';
import { getTransformerDataTypesString } from '@/util/util';
import { GetUserDefinedTransformerByIdResponse } from '@neosync/sdk';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import UpdateUserDefinedTransformerForm from './components/UpdateUserDefinedTransformerForm';

export default function UpdateUserDefinedTransformerPage({
  params,
}: PageProps) {
  const id = params?.id ?? '';
  const { account } = useAccount();

  const { data, isLoading, mutate } = useGetUserDefinedTransformersById(
    account?.id ?? '',
    id
  );

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
          leftBadgeValue={getTransformerDataTypesString(
            data?.transformer?.dataTypes ?? []
          )}
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
              {data?.transformer && (
                <UpdateUserDefinedTransformerForm
                  currentTransformer={data.transformer}
                  onUpdated={(updatedTransformer) => {
                    mutate(
                      new GetUserDefinedTransformerByIdResponse({
                        transformer: updatedTransformer,
                      })
                    );
                  }}
                />
              )}
            </div>
          </div>
        </div>
      </div>
    </OverviewContainer>
  );
}
