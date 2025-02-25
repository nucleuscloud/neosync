'use client';
import ResourceId from '@/components/ResourceId';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
import { toTitleCase } from '@/util/util';
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
  const { account } = useAccount();
  return (
    <div>
      <OverviewContainer
        Header={
          <PageHeader
            header="Settings"
            subHeadings={
              <ResourceId
                labelText={`${toTitleCase(account?.name ?? '')} - ${account?.id}`}
                copyText={account?.id ?? ''}
                onHoverText="Copy account Id"
              />
            }
          />
        }
      >
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

function NavSettings(): ReactElement<any> {
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
                : 'hover:bg-gray-200/70 hover:no-underline dark:hover:bg-gray-700'
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
    (!isSystemConfigLoading &&
      systemAppConfigData?.isAuthEnabled &&
      // filter members page if account is not a team
      account?.type === UserAccountType.TEAM) ||
    account?.type === UserAccountType.ENTERPRISE
      ? items
      : items.filter((item) => item.ref !== 'members');
  // filter temporal page if app is in neosync cloud mode
  items =
    !isSystemConfigLoading && systemAppConfigData?.isNeosyncCloud
      ? items.filter((item) => item.ref !== 'temporal')
      : items;
  // filter usage page if metrics service is not enabled
  items =
    !isSystemConfigLoading && !systemAppConfigData?.isMetricsServiceEnabled
      ? items.filter((item) => item.ref !== 'usage')
      : items;
  // filter out billing for local
  items =
    !isSystemConfigLoading && systemAppConfigData?.isNeosyncCloud
      ? items
      : items.filter((item) => item.ref !== 'billing');
  // filter out hooks if account hooks are not enabled
  items =
    !isSystemConfigLoading && !systemAppConfigData?.isAccountHooksEnabled
      ? items.filter((item) => item.ref !== 'hooks')
      : items;

  return items;
}

function getAllNavSettings(accountName: string): Item[] {
  return [
    {
      href: `/${accountName}/settings/api-keys`,
      ref: 'api-keys',
      title: 'API Keys',
    },
    {
      href: `/${accountName}/settings/temporal`,
      ref: 'temporal',
      title: 'Temporal',
    },
    {
      href: `/${accountName}/settings/members`,
      ref: 'members',
      title: 'Members',
    },
    {
      href: `/${accountName}/settings/billing`,
      ref: 'billing',
      title: 'Billing',
    },
    {
      href: `/${accountName}/settings/usage`,
      ref: 'usage',
      title: 'Usage',
    },
    {
      href: `/${accountName}/settings/hooks`,
      ref: 'hooks',
      title: 'Hooks',
    },
  ];
}

function refInPathName(ref: string, pathname: string): boolean {
  const segments = pathname.split('/');
  return segments.includes(ref);
}
