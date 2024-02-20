import ShimmeringButton from '@/components/buttons/ShimmeringButton';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';

export default function Hero(): ReactElement {
  return (
    <div className="flex flex-col gap-6 justify-center z-50 py-20">
      <div className="flex justify-center">
        <Image src="/images/nlogo2.svg" alt="logo" width={400} height={400} />
      </div>
      <div className="text-center text-gray-900 font-semibold text-3xl lg:text-6xl font-satoshi pt-10 bg-white/50">
        The Future is Synthetic Data Engineering
      </div>
      <div className="text-center text-gray-800 font-semibold text-lg font-satoshi mx-10 lg:mx-80 bg-white/50">
        Synthetic Data Engineering represents the next step in customer data
        security and privacy. Imagine having endless data, at your fingertips,
        without the security, privacy and compliance risk. When we can create
        synthetic data that is structurally and statistically exactly like your
        production data, it opens up a world of possibilities.
      </div>
      <div className="z-50 justify-center flex pt-10">
        <ShimmeringButton>
          <Link
            href="https://www.neosync.dev/blog/synthetic-data-engineering"
            target="_blank"
            className="flex flex-row items-center text-gray-300"
          >
            Read more
            <ArrowRightIcon className="ml-2 w-4 h-4" />
          </Link>
        </ShimmeringButton>
      </div>
    </div>
  );
}
