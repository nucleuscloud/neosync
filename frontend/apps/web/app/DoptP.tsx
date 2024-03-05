'use client';
import OnboardingChecklist from '@/components/checklist/OnboardingChecklist';
import { DoptProvider } from '@dopt/react';
import { ReactNode } from 'react';

interface Props {
  children: ReactNode;
}
export default function Dopt(props: Props) {
  const { children } = props;

  return (
    <DoptProvider
      userId={undefined}
      apiKey="blocks-28312c460558d766219ae7e8fd7a565d46efc6a6e1d316e48c9e15d30384ab60_MjE2MQ=="
      flowVersions={{ newuserchecklist: 3 }}
    >
      {/* {children} */}
      <OnboardingChecklist />
    </DoptProvider>
  );
}
