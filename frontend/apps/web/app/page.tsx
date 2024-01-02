import { ReactElement } from 'react';
import BaseLayout from './BaseLayout';
import AccountPage from './[account]/page';

export default function Home(): ReactElement {
  return (
    <BaseLayout>
      <AccountPage />
    </BaseLayout>
  );
}
