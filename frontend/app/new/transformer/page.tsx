'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';

export default function NewTransformer(): ReactElement {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new Transformer"
          description="Modules to anonymize, mask or generate data"
        />
      }
    >
      <div className="items-start justify-center gap-6 rounded-lg p-8 md:grid lg:grid-cols-2 xl:grid-cols-3">
        create a new transformer
      </div>
    </OverviewContainer>
  );
}
