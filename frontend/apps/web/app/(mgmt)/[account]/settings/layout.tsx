'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
import { UserAccountType } from '@neosync/sdk';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ReactElement } from 'react';

interface Item {
  href: string;
  ref: string;
  title: string;
}

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div>
      <OverviewContainer Header={<PageHeader header="Settings" />}>
        <div className="flex flex-row gap-20">
          <div className="flex flex-col gap-2 w-1/6">
            <NavSettings />
          </div>
          <div className="w-5/6">{children}</div>
        </div>
      </OverviewContainer>
    </div>
  );
}

function NavSettings(): ReactElement {
  const pathname = usePathname();
  const items = useGetNavSettings();
  return (
    <>
      {items.map((item) => {
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              'rounded p-2 text-sm',
              refInPathName(item.ref, pathname)
                ? 'bg-gray-200/70 hover:bg-gray-200/70 font-bold dark:bg-gray-700'
                : 'hover:bg-gray-200/70 hover:no-underline hover:dark:bg-gray-700'
            )}
          >
            {item.title}
          </Link>
        );
      })}
    </>
  );
}

function useGetNavSettings(): Item[] {
  const { account } = useAccount();
  const { data: systemAppConfigData, isLoading: isSystemConfigLoading } =
    useGetSystemAppConfig();

  let items = getAllNavSettings(account?.name ?? '');
  items =
    !isSystemConfigLoading &&
    systemAppConfigData?.isAuthEnabled &&
    account?.type === UserAccountType.TEAM
      ? items
      : items.filter((item) => item.ref !== 'members');
  items =
    !isSystemConfigLoading && systemAppConfigData?.isNeosyncCloud
      ? items.filter((item) => item.ref !== 'temporal')
      : items;
  return items;
}

function getAllNavSettings(accountName: string): Item[] {
  return [
    {
      href: `/${accountName}/settings/temporal`,
      ref: 'temporal',
      title: 'Temporal',
    },
    {
      href: `/${accountName}/settings/api-keys`,
      ref: 'api-keys',
      title: 'API Keys',
    },
    {
      href: `/${accountName}/settings/members`,
      ref: 'members',
      title: 'Members',
    },
  ];
}

function refInPathName(ref: string, pathname: string): boolean {
  const segments = pathname.split('/');
  return segments.includes(ref);
}
