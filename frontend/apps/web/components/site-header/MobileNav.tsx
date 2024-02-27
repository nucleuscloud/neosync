'use client';

import { ViewVerticalIcon } from '@radix-ui/react-icons';
import Link, { LinkProps } from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import * as React from 'react';

import { getMobileMainNav } from '@/app/config/mobile-nav-config';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { cn } from '@/libs/utils';
import { useTheme } from 'next-themes';
import { useAccount } from '../providers/account-provider';
import { MainLogo } from './MainLogo';
import { getPathNameHighlight } from './util';

export function MobileNav() {
  const [open, setOpen] = React.useState(false);
  const { account } = useAccount();
  const pathname = usePathname();
  const { resolvedTheme } = useTheme();

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          className="mr-2 px-0 text-base hover:bg-transparent focus-visible:bg-transparent focus-visible:ring-0 focus-visible:ring-offset-0 lg:hidden"
        >
          <ViewVerticalIcon className="h-5 w-5" />
          <span className="sr-only">Toggle Menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="pr-0">
        <MobileLink
          href="/"
          className="flex items-center"
          onOpenChange={setOpen}
        >
          <MainLogo bg={resolvedTheme === 'dark' ? 'white' : '#272F30'} />;
        </MobileLink>
        <ScrollArea className="my-4 h-[calc(100vh-8rem)] pb-10 pl-6 pt-6 w-40">
          <div className="flex flex-col space-y-3">
            {getMobileMainNav(account?.name ?? '').map(
              (item) =>
                item.href && (
                  <MobileLink
                    key={item.href}
                    href={item.href}
                    onOpenChange={setOpen}
                    className={cn(
                      'text-sm font-medium text-muted-foreground transition-colors hover:text-black dark:hover:text-white',
                      getPathNameHighlight(
                        `${item.href.split('/')[2]}`,
                        pathname
                      )
                    )}
                  >
                    {item.title}
                  </MobileLink>
                )
            )}
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}

interface MobileLinkProps extends LinkProps {
  onOpenChange?: (open: boolean) => void;
  children: React.ReactNode;
  className?: string;
}

function MobileLink({
  href,
  onOpenChange,
  className,
  children,
  ...props
}: MobileLinkProps) {
  const router = useRouter();
  return (
    <Link
      href={href}
      onClick={() => {
        router.push(href.toString());
        onOpenChange?.(false);
      }}
      className={cn(className)}
      {...props}
    >
      {children}
    </Link>
  );
}
