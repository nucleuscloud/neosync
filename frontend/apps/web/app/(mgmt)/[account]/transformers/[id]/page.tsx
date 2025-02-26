'use client';
import { use } from 'react';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { getTransformerDataTypesString } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import { createConnectQueryKey, useQuery } from '@connectrpc/connect-query';
import {
  GetUserDefinedTransformerByIdResponseSchema,
  TransformersService,
} from '@neosync/sdk';
import { useQueryClient } from '@tanstack/react-query';
import RemoveTransformerButton from './components/RemoveTransformerButton';
import UpdateTransformerForm from './components/UpdateTransformerForm';

export default function UpdateUserDefinedTransformerPage(props: PageProps) {
  const params = use(props.params);
  const id = params?.id ?? '';

  const { data, isLoading } = useQuery(
    TransformersService.method.getUserDefinedTransformerById,
    { transformerId: id },
    { enabled: !!id }
  );
  const queryclient = useQueryClient();

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
                <UpdateTransformerForm
                  currentTransformer={data.transformer}
                  onUpdated={(updatedTransformer) => {
                    const key = createConnectQueryKey({
                      schema:
                        TransformersService.method
                          .getUserDefinedTransformerById,
                      input: { transformerId: id },
                      cardinality: undefined,
                    });
                    queryclient.setQueryData(
                      key,
                      create(GetUserDefinedTransformerByIdResponseSchema, {
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
