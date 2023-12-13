'use client';
import { PageProps } from '@/components/types';
import { ReactElement } from 'react';
import AccountPage from './[account]/page';

export default function Home(props: PageProps): ReactElement {
  return <AccountPage {...props} />;
}
