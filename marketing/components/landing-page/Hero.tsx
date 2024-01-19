import { HeroandGrid } from '@/public/heroandgrid';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import { PiBookOpenText } from 'react-icons/pi';
import { Button } from '../ui/button';

export default function Hero(): ReactElement {
  return (
    <div className="flex flex-col items-center gap-10">
      <div className="text-gray-900 font-semibold lg:text-6xl text-4xl leading-tight text-center z-20 px-2 relative">
        Open Source Synthetic Data Orchestration
      </div>
      <h3 className="text-[#606060] text-md lg:text-lg font-semibold relative text-center lg:px-0 px-6">
        A developer-first way to create anonymized or synthetic data and sync it
        across all environments for testing and machine learning
      </h3>
      <div className="flex flex-col lg:flex-row lg:space-y-0 space-y-2 lg:space-x-4 z-30">
        <Button className="px-6">
          <Link href="https://github.com/nucleuscloud/neosync">
            <div className="flex flex-row gap-2">
              <GitHubLogoIcon className="mr-2 h-5 w-5" /> Get started{' '}
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
      <div className="mt-10 rounded-xl overflow-hidden">
        <div className="hidden lg:block">
          <HeroandGrid />
        </div>
        <div className="block md:hidden lg:hidden">
          <Image
            src="https://assets.nucleuscloud.com/neosync/marketingsite/mainheromobile.svg"
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
