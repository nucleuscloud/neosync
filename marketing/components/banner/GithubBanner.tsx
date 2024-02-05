'use client';
import { GitHubLogoIcon, StarFilledIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { Button } from '../ui/button';

export default function GithubBanner(): ReactElement {
  const router = useRouter();
  return (
    <div>
      <div className="top-0 hidden lg:flex flex-row gap-3 justify-center w-full h-[35px] items-center bg-gray-100 font-satoshi text-gray-900 text-sm ">
        <div>
          If you like Neosync, give it a{' '}
          <StarFilledIcon className="text-yellow-500 inline h-[20px] w-[20px]" />{' '}
          on GitHub
        </div>
        <div className="flex flex-row items-center">
          <Button
            onClick={() =>
              router.push('https://github.com/nucleuscloud/neosync')
            }
            variant="ghost"
            className="hover:bg-slate-300 "
          >
            <GitHubLogoIcon />
          </Button>
        </div>
      </div>
      <MobileBanner />
    </div>
  );
}

function MobileBanner(): ReactElement {
  const router = useRouter();
  return (
    <div className="top-0 flex md:hidden lg:hidden flex-row gap-3 justify-center w-full h-[35px] items-center bg-slate-700 font-satoshi text-gray-100">
      <div className="flex flex-row items-center">
        <StarFilledIcon className="text-yellow-500 inline h-[20px] w-[20px]" />
        <div className="text-sm pl-4">Neosync on GitHub</div>
      </div>
      <Button
        onClick={() => router.push('https://github.com/nucleuscloud/neosync')}
        variant="ghost"
        className="hover:bg-gray-300"
      >
        <GitHubLogoIcon />
      </Button>
    </div>
  );
}
