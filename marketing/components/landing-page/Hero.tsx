'use client';
import { HeroHeader } from '@/public/images/HeroSVG';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import posthog from 'posthog-js';
import { ReactElement } from 'react';
import { SiYcombinator } from 'react-icons/si';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';

/*
tailwind breakpoints
sm - 640px
md - 768px
lg - 1024px
xl - 1280px
2xl - 1536px 
*/

export default function Hero(): ReactElement {
  return (
    <div className="flex flex-col items-center gap-10 py-20 z-40">
      <Badge className="p-2 bg-white border-gray-300" variant="outline">
        <div className="flex flex-row items-center gap-2">
          <SiYcombinator className="w-5 h-5 text-orange-400 bg-white rounded-full" />
          <div>Backed by Y Combinator</div>
        </div>
      </Badge>
      <div className="text-gray-900 font-semibold text-6xl leading-tight text-center z-20 px-2 lg:w-[90%]">
        Open Source Data Anonymization and Synthetic Data Generation For
        Developers
      </div>
      <div className="text-gray-800 text-2xl font-normal text-center lg:w-[70%] xl:w-[70%] px-6 bg-white/40">
        Anonymize PII, generate synthetic data and sync environments for better
        testing, debugging and developer experience.
      </div>
      <div className="flex flex-col lg:flex-row lg:space-y-0 space-y-2 lg:space-x-4 z-30 items-center">
        <Link href="https://app.neosync.dev" target="_blank">
          <Button
            variant="default"
            className="px-6 w-[188px]"
            onClick={() =>
              posthog.capture('user click', {
                page: 'hero app sign up',
              })
            }
          >
            Try for free
            <ArrowRightIcon className="ml-2 h-4 w-4" />
          </Button>
        </Link>
        <Button
          variant="secondary"
          className="px-4 border border-gray-300 w-[188px]"
          onClick={() =>
            posthog.capture('user click', {
              page: 'hero open source button',
            })
          }
        >
          <Link href="https://github.com/nucleuscloud/neosync">
            <div className="flex flex-row items-center gap-2">
              <GitHubLogoIcon className="mr-2 h-4 w-4" /> Open Source
              <ArrowRightIcon className="h-4 w-4" />
            </div>
          </Link>
        </Button>
      </div>
      <div className="mt-10 rounded-xl overflow-hidden w-full">
        <div className="hidden md:block lg:block pt-10 w-full max-w-[1287px] ">
          <HeroHeader />
        </div>
        <div className="block md:hidden lg:hidden">
          <Image
            src="/images/mobilehero.svg"
            alt="pre"
            width="436"
            height="416"
            className="w-full"
          />
        </div>
      </div>
    </div>
  );
}
