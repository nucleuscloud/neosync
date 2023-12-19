import { ReactElement, ReactNode } from 'react';
import BaseLayout from '../BaseLayout';

export default async function InviteLayout({
  children,
}: {
  children: ReactNode;
}): Promise<ReactElement> {
  return <BaseLayout disableServerSignin>{children}</BaseLayout>;
}
