'use client';

import NextLink from 'next/link';
import { usePathname } from 'next/navigation';

import { cn } from '@/libs/utils';
import { buttonVariants } from './ui/button';

interface SidebarNavProps extends React.HTMLAttributes<HTMLElement> {
  items: {
    href: string;
    title: string;
  }[];
  buttonClassName?: string;
}

export function SidebarNav({
  className,
  items,
  buttonClassName,
  ...props
}: SidebarNavProps) {
  const pathname = usePathname();
  return (
    <nav
      className={cn(
        'flex space-x-2 lg:flex-col lg:space-x-0 lg:space-y-1',
        className
      )}
      {...props}
    >
      {items.map((item) => {
        return (
          <NextLink
            key={item.href}
            href={item.href}
            className={cn(
              buttonVariants({ variant: 'ghost' }),
              pathname === item.href
                ? 'bg-muted hover:bg-muted'
                : 'hover:bg-transparent hover:underline',
              'justify-start',
              buttonClassName
            )}
          >
            {item.title}
          </NextLink>
        );
      })}
    </nav>
  );
}
