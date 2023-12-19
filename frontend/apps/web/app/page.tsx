import { PageProps } from '@/components/types';
import { ReactElement } from 'react';
import BaseLayout from './BaseLayout';
import AccountPage from './[account]/page';

export default function Home(props: PageProps): ReactElement {
  return (
    <BaseLayout>
      <AccountPage {...props} />
    </BaseLayout>
  );
}
