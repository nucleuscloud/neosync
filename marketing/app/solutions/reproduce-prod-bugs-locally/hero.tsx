import { Button } from '@/components/ui/button';
import { KeepSync } from '@/public/images/KeepSync';
import { ArrowRightIcon, GitHubLogoIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { PiBookOpenText } from 'react-icons/pi';

export default function Hero(): ReactElement {
  return (
    <div className="flex flex-col lg:flex-row items-center lg:pb-20 mt-20 z-20 ">
      <div className="flex flex-col items-center lg:items-start gap-2 lg:gap-10 w-1/2">
        <h1 className="text-gray-900 font-semibold lg:text-6xl text-4xl leading-tight text-center lg:text-left">
          Easily reproduce Production bugs locally
        </h1>
        <h3 className="text-gray-800 text-md lg:text-lg font-semibold text-center lg:text-left lg:px-0 px-6 lg:w-[80%]">
          Anonymize senstiive Production Data to make it safe to use locally for
          the best debugging experience
        </h3>
        <div className="flex flex-col lg:flex-row lg:space-y-0 space-y-2 lg:space-x-4 pt-8">
          <Button className="px-6">
            <Link href="https://github.com/nucleuscloud/neosync">
              <div className="flex flex-row gap-2">
                <GitHubLogoIcon className="mr-2 h-5 w-5" /> Get started
                <ArrowRightIcon className="h-5 w-5" />
              </div>
            </Link>
          </Button>
          <Button variant="secondary" className="px-4">
            <Link href="https://docs.neosync.dev">
              <div className="flex flex-row items-center gap-2">
                <PiBookOpenText className="h-5 w-5" />
                Documentation
              </div>
            </Link>
          </Button>
        </div>
      </div>
      <div className="hidden lg:block pt-10 w-1/2">
        <KeepSync width={564} height={289} />
      </div>
      <div className="block md:hidden lg:hidden pt-10">
        <KeepSync width={284} height={289} />
      </div>
    </div>
  );
}
