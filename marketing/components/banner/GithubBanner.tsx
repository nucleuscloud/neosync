'use client';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

export default function GithubBanner(): ReactElement {
  return (
    <div className="w-full justify-center flex bg-slate-700 h-[35px]">
      <Link
        href="https://github.com/nucleuscloud/neosync"
        target="_blank"
        className="flex flex-row items-center top-0 gap-5 font-satoshi text-gray-100 text-sm"
      >
        Star Neosync on GitHub
        <GitHubLogoIcon />
      </Link>
    </div>
  );
}
