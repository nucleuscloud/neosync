'use client';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Menubar,
  MenubarContent,
  MenubarItem,
  MenubarMenu,
  MenubarTrigger,
} from '@/components/ui/menubar';
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
  navigationMenuTriggerStyle,
} from '@/components/ui/navigation-menu';
import { env } from '@/env';
import { cn } from '@/lib/utils';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { LucideServerCrash } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { posthog } from 'posthog-js';
import { ReactElement, ReactNode, forwardRef } from 'react';
import { AiOutlineCloudSync } from 'react-icons/ai';
import { FaDiscord } from 'react-icons/fa';
import { GiHamburgerMenu } from 'react-icons/gi';
import { GoShieldCheck } from 'react-icons/go';
import { IoTerminalOutline } from 'react-icons/io5';
import PrivateBetaButton from '../buttons/PrivateBetaButton';
import { Button } from '../ui/button';

interface NavLinks {
  title: string;
  href: string;
  description: string;
  icon?: JSX.Element;
  children: NavLinks[];
}

const links: NavLinks[] = [
  {
    title: 'About',
    href: '/about',
    description: '',
    children: [],
  },
  {
    title: 'Solutions',
    href: '/solutions',
    description: 'Solutions that Neosync can deliver',
    children: [
      {
        title: 'Unblock local development',
        href: '/solutions/unblock-local-development',
        description:
          'Self-serve anonymized and synthetic data for local development',
        children: [],
        icon: <IoTerminalOutline className="w-8 h-8" />,
      },
      {
        title: 'Fix broken staging environments',
        href: '/solutions/fix-staging-environments',
        description: 'Catch bugs faster with production-like data in staging',
        children: [],
        icon: <LucideServerCrash className="w-8 h-8" />,
      },
      {
        title: 'Keep environments up to date',
        href: '/solutions/keep-environments-in-sync',
        description:
          'Effortlessly keep environments in sync with the latest data',
        children: [],
        icon: <AiOutlineCloudSync className="w-8 h-8" />,
      },
      {
        title: 'Comply with Security and Privacy',
        href: '/solutions/security-privacy',
        description: 'Easily comply with GDPR, HIPAA, DPDP and more',
        children: [],
        icon: <GoShieldCheck className="w-8 h-8" />,
      },
    ],
  },
  {
    title: 'Docs',
    href: 'https://docs.neosync.dev',
    description: '',
    children: [],
  },
  {
    title: 'Blog',
    href: `${env.NEXT_PUBLIC_APP_URL}/blog`,
    description: '',
    children: [],
  },
  {
    title: '',
    href: 'https://github.com/nucleuscloud/neosync',
    description: '',
    icon: <GitHubLogoIcon className="h-4 w-4" />,
    children: [],
  },
  {
    title: '',
    href: 'https://discord.com/invite/MFAMgnp4HF',
    description: '',
    icon: <FaDiscord className=" h-5 w-5" />,
    children: [],
  },
];

export default function TopNav(): ReactElement {
  return (
    <div className="flex items-center justify-between px-5 sm:px-10 md:px-20 lg:px-40 max-w-[2000px] mx-auto py-4 z-50 w-full bg-[#FFFFFF]">
      <div>
        <Link href="/" className="flex items-center">
          <Image
            src="https://assets.nucleuscloud.com/neosync/newbrand/logo_text_light_mode.svg"
            alt="NeosyncLogo"
            className="w-5 object-scale-down"
            width="64"
            height="20"
          />
          <span className="text-gray-900 text-md lg:text-lg font-normal ml-1 font-satoshi">
            Neosync
          </span>
        </Link>
      </div>
      <div className="hidden items-center md:flex lg:flex lg:flex-row gap-4">
        <NavigationMenu>
          <NavigationMenuList>
            {links.map((link) =>
              link.children.length > 0 ? (
                <NavigationMenuItem key={link.href}>
                  <NavigationMenuTrigger>{link.title}</NavigationMenuTrigger>
                  <NavigationMenuContent>
                    <ul className="grid w-[400px] gap-3 p-4 md:w-[500px] md:grid-cols-2 lg:w-[600px]">
                      {link.children.map((sublink) => (
                        <ListItem
                          key={sublink.title}
                          title={sublink.title}
                          href={sublink.href}
                          icon={sublink.icon}
                        >
                          {sublink.description}
                        </ListItem>
                      ))}
                    </ul>
                  </NavigationMenuContent>
                </NavigationMenuItem>
              ) : (
                <NavigationMenuItem
                  key={link.href}
                  onClick={() =>
                    posthog.capture('user click', {
                      page: link.title,
                    })
                  }
                >
                  <Link href={link.href} passHref legacyBehavior>
                    <NavigationMenuLink
                      className={navigationMenuTriggerStyle()}
                      target="_blank"
                    >
                      {link.title ? link.title : link.icon}
                    </NavigationMenuLink>
                  </Link>
                </NavigationMenuItem>
              )
            )}
          </NavigationMenuList>
        </NavigationMenu>
        <div>
          <PrivateBetaButton />
        </div>
      </div>
      <MobileMenu />
    </div>
  );
}

function MobileMenu(): ReactElement {
  const router = useRouter();
  return (
    <div className="block md:hidden lg:hidden">
      <Menubar className="bg-transparent border border-gray-700 cursor-pointer">
        <MenubarMenu>
          <MenubarTrigger className="cursor-pointer data-[state=open]:bg-transparent focus:bg-transparent">
            <GiHamburgerMenu color="black" />
          </MenubarTrigger>
          <MenubarContent className="bg-white border border-gray-700 mx-2 mt-1 w-[244px] py-2">
            {links.map((link) =>
              link.children.length > 0 ? (
                <Accordion type="single" collapsible key={link.href}>
                  <AccordionItem value="solutions">
                    <AccordionTrigger className="no-underline flex justify-center text-[14px]">
                      Solutions
                    </AccordionTrigger>
                    <AccordionContent className="flex flex-col items-start gap-4 p-2 pl-8">
                      {link.children.map((sublink) => (
                        <Link
                          href={sublink.href}
                          passHref
                          legacyBehavior
                          key={sublink.href}
                        >
                          {sublink.title}
                        </Link>
                      ))}
                    </AccordionContent>
                  </AccordionItem>
                </Accordion>
              ) : (
                <MenubarItem key={link.href}>
                  <Button
                    variant="mobileNavLink"
                    className="text-gray-900 w-full"
                    onClick={() => {
                      posthog.capture('user click', {
                        page: link.title,
                      });
                    }}
                  >
                    <Link href={link.href}>
                      <div className="flex flex-row">
                        {link.title ? link.title : link.icon}
                      </div>
                    </Link>
                  </Button>
                </MenubarItem>
              )
            )}
            <div className="flex justify-center">
              <PrivateBetaButton />
            </div>
          </MenubarContent>
        </MenubarMenu>
      </Menubar>
    </div>
  );
}

interface ListItemProps {
  title: string;
  href: string;
  icon?: JSX.Element;
  children: ReactNode;
  className?: string;
}

const ListItem = forwardRef<HTMLAnchorElement, ListItemProps>(
  ({ className, title, href, icon, children }, ref) => {
    return (
      <li>
        <NavigationMenuLink asChild>
          <a
            ref={ref}
            href={href}
            className={cn(
              'block select-none space-y-1 rounded-md p-3 leading-none no-underline outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground',
              className
            )}
            onClick={() =>
              posthog.capture('user click', {
                page: title,
              })
            }
          >
            <div className="flex flex-row items-center gap-4 hover:text-blue-500">
              {icon && icon}
              <div className="flex flex-col gap-2">
                <div className="text-sm font-medium leading-none">{title}</div>
                <p className="line-clamp-2 text-sm leading-snug text-muted-foreground">
                  {children}
                </p>
              </div>
            </div>
          </a>
        </NavigationMenuLink>
      </li>
    );
  }
);

ListItem.displayName = 'ListItem';
