'use client';
import Image from 'next/image';
import { ReactElement } from 'react';

export default function Transformers(): ReactElement {
  const transformers = [
    {
      name: 'Pre-built Transformers',
      description: 'Choose from our library of pre-built transformers',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/pbsvg.svg"
          alt="pre"
          width="436"
          height="416"
          className="w-full"
        />
      ),
    },
    {
      name: 'Custom Transformers',
      description:
        'Create your own transformer for when you need bespoke logic ',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/customtran.svg"
          alt="pre"
          width="436"
          height="416"
          className="w-full"
        />
      ),
    },
  ];
  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="flex flex-col  pt-5 lg:pt-40 bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto ">
        <div className="text-gray-900 font-semibold  text-2xl lg:text-5xl font-satoshi">
          Fully customizable Transformers
        </div>
        <div className="flex flex-col mt-10 gap-4 border border-gray-400 bg-white p-6 rounded-xl shadow-lg">
          <div className="text-xl text-gray-700  font-satoshi font-semibold lg:w-[60%]">
            Anonymize, mask or generate data using a transformer from our
            library of pre-built tranformers or write your own transformation
            logic in code.
          </div>
          <div className="flex flex-col lg:flex-row space-y-4 lg:space-y-0 items-center justify-center gap-x-20 pt-10">
            {transformers.map((item) => (
              <div
                key={item.name}
                className="justify-center border border-gray-400 bg-white shadow-lg rounded-xl p-6"
              >
                <div className="border border-gray-900 overflow-hidden rounded-lg shadow-lg lg:w-[436px]">
                  {item.image}
                </div>
                <div className="text-xl text-gray-800 font-satoshi font-semibold pt-10">
                  {item.name}
                </div>
                <div className=" text-gray-600 font-satoshi pt-4 ">
                  {item.description}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
