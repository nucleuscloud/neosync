'use client';
import Image from 'next/image';
import { ReactElement } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';

export default function Transformers(): ReactElement {
  return (
    <div className="bg-[#F5F5F5]">
      <div className="flex flex-col pt-5 lg:pt-40 bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto ">
        <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
          The Most Flexible Way to Transform Your Data
        </div>
        <div className="text-lg text-gray-600 font-satoshi font-semibold pt-10 lg:px-60 text-center">
          Transformers anonymize existing data or generate new synthetic data.
          Choose from 35+ pre-built Transformers or write your own transformer
          in code.
        </div>
        <div className="pt-10">
          <Tabs defaultValue="system">
            <div className="flex justify-center">
              <TabsList className=" justify-center">
                <div className="border border-gray-500  rounded-lg p-2 justify-center flex lg:flex-row flex-col">
                  <TabsTrigger
                    value="system"
                    className="data-[state=active]:bg-gray-900 data-[state=active]:text-gray-100"
                  >
                    System Transformers
                  </TabsTrigger>
                  <TabsTrigger
                    value="udf"
                    className="data-[state=active]:bg-gray-900 data-[state=active]:text-gray-100"
                  >
                    User-Defined Transformers
                  </TabsTrigger>
                </div>
              </TabsList>
            </div>
            <TabsContent value="system" className="pt-10">
              <div>
                <Image
                  src="https://assets.nucleuscloud.com/neosync/marketingsite/systemTransformers.png"
                  alt="st"
                  width="1900"
                  height="1500"
                  className="rounded-xl border border-gray-400 shadow-xl"
                />
              </div>
            </TabsContent>
            <TabsContent value="udf" className="justify-center flex pt-10">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/customTransformersNew.png"
                alt="ct"
                width="1700"
                height="1200"
                className="rounded-xl border border-gray-400 shadow-xl"
              />
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>
  );
}
