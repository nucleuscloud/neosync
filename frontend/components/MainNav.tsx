'use client';

import Link from 'next/link';

import { cn } from '@/libs/utils';
import { usePathname } from 'next/navigation';
import Logo from './Logo';

export function MainNav({
  className,
  ...props
}: React.HTMLAttributes<HTMLElement>) {
  const pathname = usePathname();

  const highlightPathName = (href: string): boolean => {
    if (href === '/' && pathname === '/') {
      return true;
    }
    return href !== '/' && pathname.includes(href);
  };
  return (
    <div className="mr-4 hidden md:flex">
      <Link href="/" className="mr-6 flex items-center space-x-2">
        <Logo className="" />
        {/* <span className="hidden font-bold sm:inline-block">
          {siteConfig.name}
        </span> */}
      </Link>
      <nav
        className={cn('flex items-center space-x-4 lg:space-x-6', className)}
        {...props}
      >
        <Link
          href="/"
          className={cn(
            'text-sm font-medium transition-colors hover:text-primary',
            highlightPathName('/') ? 'text-foreground' : 'text-foreground/60'
          )}
        >
          Overview
        </Link>
        <Link
          href="/jobs"
          className={cn(
            'text-sm font-medium text-muted-foreground transition-colors hover:text-primary',
            highlightPathName('/job') ? 'text-foreground' : 'text-foreground/60'
          )}
        >
          Jobs
        </Link>
        <Link
          href="/runs"
          className={cn(
            'text-sm font-medium text-muted-foreground transition-colors hover:text-primary',
            highlightPathName('/run') ? 'text-foreground' : 'text-foreground/60'
          )}
        >
          Runs
        </Link>
        <Link
          href="/transformers"
          className={cn(
            'text-sm font-medium text-muted-foreground transition-colors hover:text-primary',
            highlightPathName('/transformer')
              ? 'text-foreground'
              : 'text-foreground/60'
          )}
        >
          Transformers
        </Link>
        <Link
          href="/connections"
          className={cn(
            'text-sm font-medium text-muted-foreground transition-colors hover:text-primary',
            highlightPathName('connection')
              ? 'text-foreground'
              : 'text-foreground/60'
          )}
        >
          Connections
        </Link>

        <Link
          href="/settings"
          className={cn(
            'text-sm font-medium text-muted-foreground transition-colors hover:text-primary',
            highlightPathName('/settings')
              ? 'text-foreground'
              : 'text-foreground/60'
          )}
        >
          Settings
        </Link>
      </nav>
    </div>
  );
}
