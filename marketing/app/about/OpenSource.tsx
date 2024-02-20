import { ArrowRightIcon, GitHubLogoIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

export default function OpenSource(): ReactElement {
  return (
    <div className="flex flex-col gap-10 justify-center">
      <div className="text-gray-200 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Open Source Commitment
      </div>
      <div className="text-center text-gray-200 font-normal text-md font-satoshi mx-6 lg:mx-40">
        We started Neosync as an open source company because we believe that
        companies should own their data. Especially their most sensitive data.
        Data privacy and security is a basic right that every company should
        protect which is why we licensed Neosync under an MIT license.
        We&apos;re fully committed to our open source product and guarantee that
        we will never take it down.
      </div>
      <div className="flex justify-center ">
        <Link href="https://github.com/nucleuscloud/neosync" target="_blank">
          <div className="flex flex-row items-center gap-2 border border-gray-600 px-4  py-2 rounded-xl bg-transparent/70 shadow-lg">
            <GitHubLogoIcon className="h-4 w-4 text-gray-100" />
            <div className="text-gray-200">Check out our Github</div>
            <ArrowRightIcon className=" ml-2h-4 w-4 text-gray-100" />
          </div>
        </Link>
      </div>
    </div>
  );
}
