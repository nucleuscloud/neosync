'use client';
import { buttonVariants } from '@/components/ui/button';
import { useGetAuthEnabled } from '@/libs/hooks/useGetAuthEnabled';
import { cn } from '@/libs/utils';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { HTMLAttributes, ReactElement } from 'react';

interface Item {
  href: string;
  title: string;
}

export function getNavSettings(accountName: string): Item[] {
  return [
    {
      href: `/${accountName}/settings`,
      title: 'Overview',
    },
    {
      href: `/${accountName}/settings/temporal`,
      title: 'Temporal',
    },
    {
      href: `/${accountName}/settings/api-keys`,
      title: 'API Keys',
    },
  ];
}

interface Props extends HTMLAttributes<HTMLElement> {
  items: Item[];
  buttonClassName?: string;
}
export default function SubNav({
  className,
  items,
  buttonClassName,
  ...props
}: Props): ReactElement {
  const pathname = usePathname();
  const authEnabled = useGetAuthEnabled();

  const filteredItems = authEnabled
    ? items
    : items.filter((item) => item.title !== 'Overview');

  return (
    <nav
      className={cn(
        'flex flex-col lg:flex-row gap-2 lg:gap-y-0 lg:gap-x-2',
        className
      )}
      {...props}
    >
      {filteredItems.map((item) => {
        if (item.href) {
        }
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              buttonVariants({ variant: 'outline' }),
              pathname === item.href
                ? 'bg-muted hover:bg-muted'
                : 'hover:bg-transparent hover:underline',
              'justify-start',
              buttonClassName,
              'px-6'
            )}
          >
            {item.title}
          </Link>
        );
      })}
    </nav>
  );
}
