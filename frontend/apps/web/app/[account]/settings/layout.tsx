'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
import { UserAccountType } from '@neosync/sdk';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

interface Item {
  href: string;
  ref: string;
  title: string;
}

function getNavSettings(accountName: string): Item[] {
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

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { account } = useAccount();
  const pathname = usePathname();
  const { data: systemAppConfigData, isLoading: isSystemConfigLoading } =
    useGetSystemAppConfig();
  const items = getNavSettings(account?.name ?? '');

  const filteredItems =
    !isSystemConfigLoading &&
    systemAppConfigData?.isAuthEnabled &&
    account?.type === UserAccountType.TEAM
      ? items
      : items.filter((item) => item.title !== 'Members');

  return (
    <div>
      <OverviewContainer Header={<PageHeader header="Settings" />}>
        <div className="flex flex-row gap-20">
          <div className="flex flex-col gap-2 w-1/6">
            {filteredItems.map((item) => {
              if (item.href) {
              }
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
          </div>
          <div className="w-5/6">{children}</div>
        </div>
      </OverviewContainer>
    </div>
  );
}

function refInPathName(ref: string, pathname: string): boolean {
  const segments = pathname.split('/');
  return segments.includes(ref);
}
