'use client';
import { GitHubLogoIcon, StarFilledIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

export default function GithubBanner(): ReactElement {
  return (
    <div className="w-full justify-center flex bg-slate-700 h-[35px]">
      <Link
        href="https://github.com/nucleuscloud/neosync"
        target="_blank"
        className="flex flex-row items-center top-0 gap-5 font-satoshi text-gray-100"
      >
        <StarFilledIcon className="text-yellow-500 inline h-[20px] w-[20px]" />
        <div className="text-sm pl-4">Neosync on GitHub</div>
        <GitHubLogoIcon />
      </Link>
    </div>
  );
}
