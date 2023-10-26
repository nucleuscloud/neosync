'use client';
import PageHeader from '@/components/headers/PageHeader';
import { UpdateConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { CustomTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement } from 'react';
import { handleTransformerMetadata } from '../../EditTransformerOptions';

interface TransformerComponent {
  name: string;
  summary?: ReactElement;
  body: ReactElement;
  header: ReactElement;
}

interface GetTransformerComponentDetailsProps {
  transformer?: CustomTransformer;
  onSaved(updatedConnResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
  extraPageHeading?: ReactElement;
}

export function getTransformerComponentDetails(
  props: GetTransformerComponentDetailsProps
): TransformerComponent {
  const { transformer } = props;

  return {
    name: transformer?.name ?? 'Undefined',
    summary: (
      <div>
        <p>No summary found.</p>
      </div>
    ),
    header: (
      <PageHeader
        header={`${transformer?.name} Transformer`}
        description={
          handleTransformerMetadata(transformer?.description).description
        }
      />
    ),
    body: (
      //   <div>{handleTransformerForm(transformer?.value ?? 'passthrough')}</div>
      <div>this is the {transformer?.name} component</div>
    ),
  };
}
