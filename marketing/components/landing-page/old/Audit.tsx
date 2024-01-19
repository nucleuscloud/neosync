'use client';
import GradientTag from '@/components/GradientTag';
import Image from 'next/image';
import { ReactElement } from 'react';

export default function AuditSection(): ReactElement {
  return (
    <div className="flex flex-col pt-20 lg:pt-40 items-center">
      <GradientTag tagValue={'Security and Compliance'} />
      <div className="pt-10 text-center z-3 text-gray-900 font-semibold  text-2xl lg:text-5xlfont-satoshi z-10">
        Access Controls and Audit Logs for Compliance and Security
      </div>
      <div className="flex flex-col lg:flex-row px-1 lg:px-5 pt-10 space-y-5 lg:space-x-5 lg:space-y-0 z-1 items-center z-10">
        <div className="bg-gradient-to-tr from-[#0F0F0F] to-[#191919] border-[1px] border-gray-700  rounded-xl text-gray-200 lg:w-[60%]">
          <div className="flex flex-col p-5">
            <div className="flex flex-col text-left font-semibold text-gray-200 text-xl">
              Fine grained Access Controls
            </div>
            <div className="text-left text-lg text-gray-400 font-semibold">
              Access control govern which members have access to create jobs,
              transformers, invite members and more.
            </div>
            <div className="border border-gray-700 rounded-xl mt-5 overflow-hidden">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/rbac.png"
                alt="pre"
                width="549"
                height="554"
                className="w-full"
              />
            </div>
          </div>
        </div>
        <div className="bg-gradient-to-tr from-[#0F0F0F] to-[#191919] border-[1px] border-gray-700 rounded-xl text-gray-200 lg:w-[60%]">
          <div className="flex flex-col p-5">
            <div className="flex flex-col text-left font-semibold text-gray-200 text-xl">
              Audit Logs for Compliance
            </div>
            <div className="text-left text-lg text-gray-400 font-semibold">
              Stay in control of your data and track every event in the Audit
              log. Export it to stay up to date with compliance.
            </div>
            <div className="border border-gray-700 rounded-xl overflow-hidden mt-5">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/audittrimmed.png"
                alt="pre"
                className="w-full"
                width="549"
                height="554"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
