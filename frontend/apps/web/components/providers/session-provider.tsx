'use client';

import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { isPast, parseISO } from 'date-fns';
import { Session } from 'next-auth';
import {
  SessionProvider as NextAuthSessionProvider,
  signIn,
} from 'next-auth/react';
import { ReactNode } from 'react';
import { Skeleton } from '../ui/skeleton';

interface Props {
  children: ReactNode;
  session: Session | null;
}

export function SessionProvider({ children, session }: Props) {
  const { data, isLoading } = useGetSystemAppConfig();
  if (isLoading) {
    return <Skeleton />;
  }
  if (data?.isAuthEnabled && !isSessionValid(session)) {
    signIn(data.signInProviderId);
    return <Skeleton />;
  }
  return (
    <NextAuthSessionProvider session={session}>
      {children}
    </NextAuthSessionProvider>
  );
}

function isSessionValid(session: Session | null): boolean {
  if (!session) {
    return false;
  }
  const expiryDate = parseISO(session.expires);
  return !isPast(expiryDate);
}
