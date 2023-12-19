import { ReactElement, ReactNode } from 'react';
import BaseLayout from '../BaseLayout';

export default async function InviteLayout({
  children,
}: {
  children: ReactNode;
}): Promise<ReactElement> {
  // Server Signin is disabled for the invite page due to inability to access path or search params on the server
  // Without this, the signin redirect url is set to the root url instead of /invite?token=<token>
  return <BaseLayout disableServerSignin>{children}</BaseLayout>;
}
