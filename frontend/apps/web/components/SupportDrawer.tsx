import { siteConfig } from '@/app/config/site';
import {
  ArrowTopRightIcon,
  DiscordLogoIcon,
  GitHubLogoIcon,
} from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { PiBookOpenText } from 'react-icons/pi';

export default function SupportDrawer(): ReactElement {
  const supportLinks = [
    {
      name: 'Documentation',
      href: siteConfig.links.docs,
      icon: <PiBookOpenText className="h-4 w-4" />,
      description: 'Guides, useful information and more. ',
    },
    {
      name: 'Github',
      href: siteConfig.links.github,
      icon: <GitHubLogoIcon />,
      description: 'Check out our source code.',
    },
    {
      name: 'Discord',
      href: siteConfig.links.discord,
      icon: <DiscordLogoIcon />,
      description: 'Ask us a question directly!',
    },
  ];

  return (
    <div className="flex flex-col gap-3 pt-10">
      {supportLinks.map((link) => (
        <Link
          key={link.name}
          href={link.href}
          target="_blank"
          rel="noreferrer"
          className="border border-gray-300 dark:border-gray-700 bg-transparent shadow-xs hover:bg-accent hover:text-accent-foreground rounded-md px-8 py-2"
        >
          <div className="flex flex-row justify-between items-center">
            <div className="flex flex-col gap-1">
              <div className="flex flex-row items-center gap-2">
                <div>{link.icon}</div>
                <div>{link.name}</div>
              </div>
              <div className="text-sm text-gray-500">{link.description}</div>
            </div>
            <ArrowTopRightIcon />
          </div>
        </Link>
      ))}
    </div>
  );
}
