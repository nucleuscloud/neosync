'use client';
import PageHeader from '@/components/headers/PageHeader';
import { UpdateConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
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

  // switch (transformer?.title) {
  //   case 'Passthrough':
  //     return {
  //       name: transformer.title,
  //       summary: (
  //         <div>
  //           <p>No summary found.</p>
  //         </div>
  //       ),
  //       header: (
  //         <PageHeader
  //           header="Unknown Transformer"
  //           description="Update this transformer"
  //         />
  //       ),
  //       body: (
  //         <div>
  //           No connection component found for: (
  //           {transformer?.title ?? 'unknown name'})
  //         </div>
  //       ),
  //     };
  //   default:
  //     return {
  //       name: 'Invalid Connection',
  //       summary: (
  //         <div>
  //           <p>No summary found.</p>
  //         </div>
  //       ),
  //       header: (
  //         <PageHeader
  //           header="Unknown Transformer"
  //           description="Update this transformer"
  //         />
  //       ),
  //       body: (
  //         <div>
  //           No connection component found for: (
  //           {transformer?.title ?? 'unknown name'})
  //         </div>
  //       ),
  //     };
  // }

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
