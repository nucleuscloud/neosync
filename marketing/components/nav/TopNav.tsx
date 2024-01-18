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
import { env } from '@/env';
import { FireMixpanel } from '@/lib/mixpanel';
import { ArrowRightIcon, GitHubLogoIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { FaDiscord } from 'react-icons/fa';
import { GiHamburgerMenu } from 'react-icons/gi';
import PrivateBetaForm from '../buttons/PrivateBetaForm';
import { Button } from '../ui/button';

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
        <div>
          <Button
            variant="navLink"
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
        </div>
        <div className="flex items-center">
          <Button
            variant="navLink"
            onClick={() => {
              FireMixpanel('docs', {
                source: 'top-nav',
                type: 'docs section',
              });
            }}
          >
            <Link href="https://docs.neosync.dev">
              <div className="flex flex-row">Docs</div>
            </Link>
          </Button>
        </div>
        <div className="flex items-center">
          <Button
            variant="navLink"
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
        </div>
        <div>
          <Button
            variant="navLink"
            onClick={() => {
              FireMixpanel('github button', {
                source: 'top-nav',
                type: 'github button',
              });
            }}
          >
            <Link href="https://github.com/nucleuscloud/neosync">
              <div className="flex flex-row items-center">
                <GitHubLogoIcon className="h-4 w-4" />
              </div>
            </Link>
          </Button>
        </div>
        <div>
          <Button variant="navLink">
            <Link href="https://discord.gg/UVmPTzn7dV">
              <FaDiscord className=" h-5 w-5" />
            </Link>
          </Button>
        </div>
        <div>
          <Dialog>
            <DialogTrigger asChild>
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
                  Want to use Neosync but don&apos;t want to host it yourself?
                  Sign up for the private beta of Neosync Cloud.
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
