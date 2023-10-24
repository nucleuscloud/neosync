'use client';
import PageHeader from '@/components/headers/PageHeader';
import { UpdateConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement } from 'react';

interface TransformerComponent {
  name: string;
  summary?: ReactElement;
  body: ReactElement;
  header: ReactElement;
}

interface GetTransformerComponentDetailsProps {
  transformer?: Transformer;
  onSaved(updatedConnResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
  extraPageHeading?: ReactElement;
}

export function getTransformerComponentDetails(
  props: GetTransformerComponentDetailsProps
): TransformerComponent {
  const { transformer } = props;

  switch (transformer?.value) {
    case 'Passthrough':
      return {
        name: transformer.value,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header={`${transformer.value} Transformer`}
            description="This transformer is a passthrough."
          />
        ),
        body: <div>{transformer?.value} transformer for right now</div>,
      };
    default:
      return {
        name: 'Invalid Connection',
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header="Unknown Transformer"
            description="Update this transformer"
          />
        ),
        body: (
          <div>
            No connection component found for: (
            {transformer?.value ?? 'unknown name'})
          </div>
        ),
      };
  }
}
