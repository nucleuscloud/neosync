'use client';

import { Session } from 'next-auth';
import {
  SessionProvider as NextAuthSessionProvider,
  signIn,
} from 'next-auth/react';
import { ReactNode, useEffect } from 'react';

interface Props {
  children: ReactNode;
  session: Session | null;
  isAuthEnabled: boolean;
}

export function SessionProvider({ children, session, isAuthEnabled }: Props) {
  useEffect(() => {
    if (!session && isAuthEnabled) {
      signIn();
    }
  }, [session?.expires, isAuthEnabled]);
  return (
    <NextAuthSessionProvider session={session}>
      {children}
    </NextAuthSessionProvider>
  );
}
