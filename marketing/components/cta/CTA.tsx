'use client';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import posthog from 'posthog-js';
import { ReactElement } from 'react';
import SignupForm from '../buttons/SignupForm';
import { Button } from '../ui/button';

export default function CTA(): ReactElement {
  return (
    <div className=" bg-gradient-to-r from-slate-50 to-zinc-300 border border-gray-400 shadow-xl rounded-xl flex flex-col align-center space-y-6 py-10 justify-center">
      <div className="flex flex-row gap-4 items-center px-2 lg:px-10">
        <div className="flex flex-col gap-8 lg:pl-10">
          <div className="text-gray-900 text-2xl lg:text-4xl font-satoshi font-bold md:text-center lg:text-left">
            Start today with Neosync
          </div>
          <div className="text-md text-gray-800 font-satoshi font-semibold md:px-10 lg:px-0 lg:w-[80%] md:text-center lg:text-left">
            Get started with Neosync Cloud or easily deploy open source Neosync
            using a Helm chart or Docker Componse file.
          </div>
          <div className="flex flex-col lg:flex-row items-center gap-6">
            <div>
              <Dialog>
                <DialogTrigger asChild>
                  <Button
                    variant="default"
                    className="w-[188px]"
                    onClick={() =>
                      posthog.capture('user click', {
                        page: 'cta neosync cloud button',
                      })
                    }
                  >
                    Neosync Cloud <ArrowRightIcon className="ml-2 h-5 w-8 " />
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-lg bg-white p-6 shadow-xl">
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
                      Get access to Neosync Cloud
                    </DialogTitle>
                    <DialogDescription className="pt-6 text-gray-900 text-md text-center">
                      Want to use Neosync but don&apos;t want to host it
                      yourself? Let&apos;s chat.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="flex items-center space-x-2">
                    <SignupForm />
                  </div>
                </DialogContent>
              </Dialog>
            </div>
            <Button
              className="px-6 w-[188px]"
              variant="secondary"
              onClick={() =>
                posthog.capture('user click', {
                  page: 'cta open source button',
                })
              }
            >
              <Link href="https://github.com/nucleuscloud/neosync">
                <div className="flex flex-row gap-2">
                  <GitHubLogoIcon className="mr-2 h-5 w-5" /> Open Source
                </div>
              </Link>
            </Button>
          </div>
        </div>
        <div className="hidden lg:flex">
          <Image
            src="https://assets.nucleuscloud.com/neosync/marketingsite/neosync-3d.svg"
            alt="NeosyncLogo"
            width="499"
            height="416"
          />
        </div>
      </div>
    </div>
  );
}
