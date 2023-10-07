'use client';
import PageHeader from '@/components/headers/PageHeader';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { ReactElement } from 'react';
import EmailTransformerForm from './EmailTransformerForm';
import FirstNameTransformerForm from './FirstNameForm';

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

  console.log('tra', transformer);

  switch (transformer?.title) {
    case 'Email':
      return {
        name: transformer.title,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header={`${transformer.title} Transformer`}
            description="This transformer is a passthrough."
          />
        ),
        body: <div>{handleTransformerForm(transformer)}</div>,
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
            {transformer?.title ?? 'unknown name'})
          </div>
        ),
      };
  }
}

export function handleTransformerForm(transformer: Transformer): ReactElement {
  switch (transformer.title) {
    case 'Email':
      return <EmailTransformerForm transformer={transformer} />;
    case 'First Name':
      return <FirstNameTransformerForm transformer={transformer} />;
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}
