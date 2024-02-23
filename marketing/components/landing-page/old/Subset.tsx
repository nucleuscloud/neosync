'use client';
import Image from 'next/image';
import { ReactElement } from 'react';

export default function Subset(): ReactElement {
  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" flex flex-col space-y-20 bg-gradient-to-tr from-[#0F0F0F] to-[#262626] rounded-2xl py-12 px-6 shadow-lg">
          <div className="text-gray-100 font-semibold  text-2xl lg:text-4xl font-satoshi pt-10">
            Powerful subsetting
          </div>
          <div className="text-lg text-gray-400 font-satoshi font-light pt-8 lg:w-[70%]">
            Subsetting allows you to filter your source database so that you can
            easily reproduce bugs and data errors and shrink your production
            database so that it fits locally. Neosync maintains full referential
            integrity of your data automatically.
          </div>
          <div className="hidden lg:justify-center lg:flex">
            {/* <SubsetAnimation /> */}
          </div>
          <div className="block md:hidden lg:hidden ">
            <Image
              src="https://assets.nucleuscloud.com/neosync/marketingsite/customfilters.png"
              alt="pre"
              width="436"
              height="416"
              className="w-full"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
