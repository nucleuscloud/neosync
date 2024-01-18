import Image from 'next/image';
import { ReactElement } from 'react';

export default function EngineeringTeams(): ReactElement {
  return (
    <div className="bg-[#F5F5F5] relative">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
          Built for Engineering Teams
        </div>
        <div className="text-lg text-gray-600 font-satoshi font-semibold pt-10 px-60 text-center">
          From Local, to Stage to CI, Neosync has APIs, SDKs and a CLI to fit
          into every workflow.
        </div>
        <div className="p-6 rounded-xl mt-10 mx-40 border border-gray-400 bg-white shadow-xl">
          <div className="rounded-xl flex justify-center">
            <Image
              src="https://assets.nucleuscloud.com/neosync/marketingsite/codedotgrid.png"
              alt="pre"
              width="900"
              height="642"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
