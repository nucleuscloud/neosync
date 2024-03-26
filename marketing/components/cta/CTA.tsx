'use client';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import posthog from 'posthog-js';
import { ReactElement } from 'react';
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
            <Link href="https://app.neosync.dev" target="_blank">
              <Button
                variant="default"
                className="px-6 w-[188px]"
                onClick={() =>
                  posthog.capture('user click', {
                    page: 'cta app sign up',
                  })
                }
              >
                Neosync Cloud <ArrowRightIcon className="ml-2 h-4 w-4" />
              </Button>
            </Link>
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
