import AccountPageRedirect from '@/components/AccountPageRedirect';
import { ReactElement, ReactNode } from 'react';
import BaseLayout from '../BaseLayout';

export default async function AccountLayout({
  children,
}: {
  children: ReactNode;
}): Promise<ReactElement> {
  return (
    <BaseLayout>
      <AccountPageRedirect>{children}</AccountPageRedirect>
    </BaseLayout>
  );
}
