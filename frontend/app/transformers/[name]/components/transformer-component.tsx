'use client';
import PageHeader from '@/components/headers/PageHeader';
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
  // onSaved(updatedConnResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
  extraPageHeading?: ReactElement;
}

export function getTransformerComponentDetails(
  props: GetTransformerComponentDetailsProps
): TransformerComponent {
  const { transformer } = props;

  switch (transformer?.name) {
    case 'email':
      return {
        name: transformer.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header={`${transformer.name} Transformer`}
            description="This transformer is a passthrough."
          />
        ),
        body: <div>{transformer?.name} transformer for right now</div>,
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
            {transformer?.name ?? 'unknown name'})
          </div>
        ),
      };
  }
}
