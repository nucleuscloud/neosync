'use client';
import { Button } from '@/components/ui/button';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { BiSolidQuoteLeft, BiSolidQuoteRight } from 'react-icons/bi';

export default function ProblemStmt(): ReactElement {
  const router = useRouter();
  return (
    <div className="bg-[#F5F5F5] mt-20">
      <div className="flex flex-col justify-center px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <Button variant="link" className="pt-20 text-gray-400">
          <Link href="/about">
            <div className="flex flex-row items-center space-x-2 w-full justify-center z-20 hover:underline hover:text-gray-800">
              <div className="font-normal text-gray-800 py-2">
                About us & our Investors
              </div>
              <ArrowRightIcon className="w-3 h-3 text-gray-800" />
            </div>
          </Link>
        </Button>
        <div className="p-4 lg:p-10 lg:m-10 mt-10 lg:mt-20 border-2 border-gray-700 rounded-xl shadow-lg z-1 bg-[#262626] z-10 ">
          <div className="flex flex-col">
            <BiSolidQuoteLeft className="w-10 h-10 text-blue-200" />
            <div className="font-normal text-lg text-gray-300 font-satoshi font-20px">
              Test data management is inconsistently adopted across
              organizations, with most teams copying production data for use in
              test environments...this traditional approach is increasingly at
              odds with requirements for efficiency, privacy and security.
            </div>
            <div className="flex justify-end">
              <BiSolidQuoteRight className="w-10 h-10 text-blue-200" />
            </div>
            <div className="flex flex-col pt-10">
              <div className="text-lg font-normal text-gray-200 font-satoshi">
                Gartner
              </div>
              <div className="text-md font-light text-gray-400 font-satoshi">
                Hype Cycle for Agile & DevOps, 2022
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
