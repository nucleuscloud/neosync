import { siteConfig } from '@/app/config/site';
import { buttonVariants } from '@/components/ui/button';
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

export default function SupportDrawer(): ReactElement {
  const iconButtonClassNames = cn(
    buttonVariants({
      variant: 'ghost',
    }),
    'p-0 px-2'
  );
  return (
    <div>
      <div className="flex flex-1 justify-end items-center space-x-2">
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
        <Link href={siteConfig.links.discord} target="_blank" rel="noreferrer">
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
      </div>
    </div>
  );
}
