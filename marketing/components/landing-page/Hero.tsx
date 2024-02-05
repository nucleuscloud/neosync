import { HeroHeader } from '@/public/images/MainHero';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import { PiBookOpenText } from 'react-icons/pi';
import PrivateBetaButton from '../buttons/PrivateBetaButton';
import { Button } from '../ui/button';

export default function Hero(): ReactElement {
  return (
    <div className="flex flex-col items-center gap-10 py-20 z-40">
      <div className="text-gray-900 font-semibold lg:text-6xl text-4xl leading-tight text-center z-20 px-2 relative">
        Open Source Synthetic Data Orchestration
      </div>
      <h3 className="text-gray-800 text-md lg:text-lg font-semibold relative text-center lg:px-0 px-6">
        A developer-first way to create anonymized or synthetic data and sync it
        across all environments for stage, local and CI testing
      </h3>
      <div className="flex flex-col lg:flex-row lg:space-y-0 space-y-2 lg:space-x-4 z-30">
        <PrivateBetaButton />

        <Button variant="secondary" className="px-4">
          <Link href="https://docs.neosync.dev">
            <div className="flex flex-row items-center gap-2">
              <PiBookOpenText className="mr-2 h-5 w-5" /> Documentation
              <ArrowRightIcon className="h-4 w-4" />
            </div>
          </Link>
        </Button>
      </div>
      <div className="mt-10 rounded-xl overflow-hidden">
        <div className="hidden lg:block pt-10">
          <HeroHeader />
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
