'use client';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
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
import { FireMixpanel } from '@/lib/mixpanel';
import { cn } from '@/lib/utils';
import { ArrowRightIcon, GitHubLogoIcon } from '@radix-ui/react-icons';
import { LucideServerCrash } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement, ReactNode, forwardRef } from 'react';
import { AiOutlineCloudSync } from 'react-icons/ai';
import { FaDiscord } from 'react-icons/fa';
import { GiHamburgerMenu } from 'react-icons/gi';
import { GoShieldCheck } from 'react-icons/go';
import { IoTerminalOutline } from 'react-icons/io5';
import PrivateBetaForm from '../buttons/PrivateBetaForm';
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
        href: '/unblock-local-development',
        description:
          'Self-serve anonymized and synthetic data for local development',
        children: [],
        icon: <IoTerminalOutline className="w-8 h-8" />,
      },
      {
        title: 'Fix broken staging environments',
        href: '/fix-staging-environments',
        description: 'Catch bugs faster with production-like data in staging',
        children: [],
        icon: <LucideServerCrash className="w-8 h-8" />,
      },
      {
        title: 'Keep environments up to date',
        href: '/keep-environments-in-sync',
        description:
          'Effortlessly keep environments in sync with the latest data',
        children: [],
        icon: <AiOutlineCloudSync className="w-8 h-8" />,
      },
      {
        title: 'Comply with Security and Privacy',
        href: '/comply-security-privacy',
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
    href: 'https://discord.gg/UVmPTzn7dV',
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
      <NavigationMenu>
        <NavigationMenuList>
          {links.map((link) =>
            link.children.length > 0 ? (
              <NavigationMenuItem
                key={link.href}
                onClick={() =>
                  FireMixpanel(`${link.title}`, {
                    source: 'top-nav',
                    type: 'click',
                  })
                }
              >
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
                  FireMixpanel(`${link.title}`, {
                    source: 'top-nav',
                    type: 'click',
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
        <Dialog>
          <DialogTrigger asChild>
            <Button variant="default">
              Neosync Cloud <ArrowRightIcon className="ml-2 h-5 w-5" />
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-lg bg-white  p-6 shadow-xl">
            <DialogHeader>
              <div className="flex justify-center pt-10">
                <Image
                  src="https://assets.nucleuscloud.com/neosync/newbrand/logo_text_light_mode.svg"
                  alt="NeosyncLogo"
                  width="118"
                  height="30"
                />
              </div>
              <DialogTitle className="text-gray-900 text-2xl text-center pt-10">
                Join the Neosync Cloud Private Beta
              </DialogTitle>
              <DialogDescription className="pt-6 text-gray-900 text-md text-center">
                Want to use Neosync but don&apos;t want to host it yourself?
                Sign up for the private beta of Neosync Cloud and get an
                environment.
              </DialogDescription>
            </DialogHeader>
            <div className="flex items-center space-x-2">
              <PrivateBetaForm />
            </div>
          </DialogContent>
        </Dialog>
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
          <MenubarContent className="bg-white border border-gray-700 mx-2 mt-1">
            <MenubarItem>
              <Button
                variant="mobileNavLink"
                className="text-gray-900 w-full"
                onClick={() => {
                  FireMixpanel('About Section', {
                    source: 'top-nav',
                    type: 'about section',
                  });
                }}
              >
                <Link href="/about">
                  <div className="flex flex-row">About</div>
                </Link>
              </Button>
            </MenubarItem>
            <MenubarItem>
              <Button
                variant="navLink"
                className="text-gray-900 w-full"
                onClick={() => {
                  FireMixpanel('blog', {
                    source: 'top-nav',
                    type: 'blog section',
                  });
                }}
              >
                <Link href={`${env.NEXT_PUBLIC_APP_URL}/blog`}>
                  <div className="flex flex-row">Blog</div>
                </Link>
              </Button>
            </MenubarItem>
            <MenubarItem>
              <Button
                variant="mobileNavLink"
                className="text-gray-900 w-full"
                id="2"
                onClick={() => {
                  FireMixpanel('About Section', {
                    source: 'top-nav',
                    type: 'about section',
                  });
                  router.push('/about');
                }}
              >
                <Link href="https://github.com/nucleuscloud/neosync">
                  <div className="flex flex-row items-center">
                    <GitHubLogoIcon className=" h-4 w-4" />
                  </div>
                </Link>
              </Button>
            </MenubarItem>
            <MenubarItem>
              <Dialog>
                <DialogTrigger asChild className="w-full">
                  <Button variant="default">
                    Neosync Cloud <ArrowRightIcon className="ml-2 h-5 w-5" />
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-lg bg-black border border-gray-600 p-6">
                  <DialogHeader>
                    <DialogTitle className="text-white text-2xl">
                      Join the Neosync Cloud Private Beta
                    </DialogTitle>
                    <DialogDescription className="pt-10 text-gray-300 text-md">
                      Want to use Neosync but don&apos;t want to host it
                      yourself? Sign up for the private beta of Neosync Cloud.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="flex items-center space-x-2">
                    <PrivateBetaForm />
                  </div>
                  <DialogFooter className="sm:justify-start">
                    <DialogClose asChild>
                      <Button
                        type="button"
                        variant="ghost"
                        className="text-white hover:bg-gray-800 hover:text-white"
                      >
                        Close
                      </Button>
                    </DialogClose>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </MenubarItem>
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
          >
            <div className="flex flex-row items-center gap-4 ">
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
