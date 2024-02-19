'use client';
import { GitHubLogoIcon, StarFilledIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';

export default function GithubButton(): ReactElement {
  const router = useRouter();
  return (
    <div className="fixed bottom-20 right-20 z-50 bg-gray-100 rounded-xl border border-gray-800 shadow-xl overflow-hidden p-3">
      <Link href="https://github.com/nucleuscloud/neosync" target="_blank">
        <div className=" flex flex-row gap-2 items-center">
          <div>
            <StarFilledIcon className="text-yellow-500 inline h-[20px] w-[20px]" />{' '}
          </div>
          <div>Star on GitHub</div>
          <div>
            <GitHubLogoIcon />
          </div>
        </div>
      </Link>
    </div>
  );
}
