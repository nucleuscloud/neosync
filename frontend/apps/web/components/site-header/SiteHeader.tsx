import { getSystemAppConfig } from '@/app/api/config/config';
import { siteConfig } from '@/app/config/site';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { DiscordLogoIcon, GitHubLogoIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { PiBookOpenText } from 'react-icons/pi';
import { buttonVariants } from '../ui/button';
import AccountSwitcher from './AccountSwitcher';
import { MainNav } from './MainNav';
import { MobileNav } from './MobileNav';
import { ModeToggle } from './ModeToggle';
import { UserNav } from './UserNav';

export default function SiteHeader(): ReactElement {
  const systemAppConfig = getSystemAppConfig();
  const iconButtonClassNames = cn(
    buttonVariants({
      variant: 'ghost',
    }),
    'p-0 px-2'
  );
  return (
    <header className="supports-backdrop-blur:bg-background/60 sticky top-0 z-50 w-full border-b dark:border-b-gray-700 bg-background dark:hover:text-white backdrop-blur">
      <div className="container flex h-14 items-center">
        <MainNav />
        <MobileNav />
        <div className="flex flex-1 justify-end items-center space-x-2">
          {systemAppConfig.isAuthEnabled && <AccountSwitcher />}
          <Link href={siteConfig.links.docs} target="_blank" rel="noreferrer">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className={iconButtonClassNames}>
                    <PiBookOpenText className="h-4 w-4" />
                  </div>
                </TooltipTrigger>
                <TooltipContent side="bottom">Documentation</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </Link>
          <Link href={siteConfig.links.github} target="_blank" rel="noreferrer">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className={iconButtonClassNames}>
                    <GitHubLogoIcon />
                  </div>
                </TooltipTrigger>
                <TooltipContent side="bottom">Github</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </Link>
          <Link
            href={siteConfig.links.discord}
            target="_blank"
            rel="noreferrer"
          >
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className={iconButtonClassNames}>
                    <DiscordLogoIcon />
                  </div>
                </TooltipTrigger>
                <TooltipContent side="bottom">Discord</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </Link>
          <ModeToggle />
          <UserNav />
        </div>
      </div>
    </header>
  );
}
