'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useGetAuthEnabled } from '@/libs/hooks/useGetAuthEnabled';
import { cn } from '@/libs/utils';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

interface Item {
  href: string;
  title: string;
}

export const links: Item[] = [
  {
    href: '/settings',
    title: 'Overview',
  },
  {
    href: '/settings/temporal',
    title: 'Temporal',
  },
  {
    href: '/settings/account-api-keys',
    title: 'API Keys',
  },
  {
    href: '/settings/members',
    title: 'Members',
  },
];

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const authEnabled = useGetAuthEnabled();

  const filteredItems = authEnabled
    ? links
    : links.filter((item) => item.title !== 'Overview');

  return (
    <div>
      <OverviewContainer
        Header={<PageHeader header="Settings" />}
        containerClassName="px-12 md:px-24 lg:px-32"
      >
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
                    pathname === item.href
                      ? 'bg-gray-200 hover:bg-gray-200 font-medium'
                      : 'hover:bg-gray-200 hover:no-underline'
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
