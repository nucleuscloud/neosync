'use client';
import PageHeader from '@/components/headers/PageHeader';
import { UpdateConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { CustomTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement } from 'react';
import UpdateTransformerForm from './UpdateTransformerForm';

interface TransformerComponent {
  name: string;
  summary?: ReactElement;
  body: ReactElement;
  header: ReactElement;
}

interface GetTransformerComponentDetailsProps {
  CustomTransformer?: CustomTransformer;
  onSaved(updatedConnResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
  extraPageHeading?: ReactElement;
}

export function getTransformerComponentDetails(
  props: GetTransformerComponentDetailsProps
): TransformerComponent {
  const { CustomTransformer } = props;

  return {
    name: CustomTransformer?.name ?? 'Custom Transformer',
    summary: (
      <div>
        <p>No summary found.</p>
      </div>
    ),
    header: (
      <PageHeader header={CustomTransformer?.name ?? 'Custom Transformer'} />
    ),
    body: (
      <div>
        <UpdateTransformerForm currentTransformer={CustomTransformer} />
      </div>
    ),
  };
}
