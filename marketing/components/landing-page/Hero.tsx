import { HeroHeader } from '@/public/images/MainHero';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import { PiBookOpenText } from 'react-icons/pi';
import { SiYcombinator } from 'react-icons/si';
import PrivateBetaButton from '../buttons/PrivateBetaButton';
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
          <SiYcombinator className="w-4 h-4 text-orange-400 bg-white rounded-full" />
          <div>Backed by Y Combinator</div>
        </div>
      </Badge>
      <div className="text-gray-900 font-semibold text-6xl leading-tight text-center z-20 px-2 ">
        Synthetic Test Data For Developers
      </div>
      <div className="text-gray-800 text-lg font-semibold text-center lg:w-[70%] xl:w-full px-6">
        A developer-first way to create anonymized or synthetic data and sync it
        across all environments for stage, local and CI testing
      </div>
      <div className="flex flex-col lg:flex-row lg:space-y-0 space-y-2 lg:space-x-4 z-30 items-center">
        <PrivateBetaButton />
        <Button variant="secondary" className="px-4 border border-gray-300">
          <Link href="https://docs.neosync.dev">
            <div className="flex flex-row items-center gap-2">
              <PiBookOpenText className="mr-2 h-5 w-5" /> Documentation
              <ArrowRightIcon className="h-4 w-4" />
            </div>
          </Link>
        </Button>
      </div>
      <div className="mt-10 rounded-xl overflow-hidden w-full">
        <div className="hidden md:block lg:block pt-10 w-full max-w-[1287px] h-auto">
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
