'use client';
import {
  GitHubLogoIcon,
  StarFilledIcon,
  TwitterLogoIcon,
} from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { Button } from '../ui/button';

export default function GithubBanner(): ReactElement {
  const router = useRouter();
  return (
    <div>
      <div className="top-0 hidden lg:flex flex-row gap-3 justify-center w-full h-[35px] items-center bg-gradient-to-r from-indigo-200 to-slate-300 font-satoshi">
        <div>
          If you like Neosync, give it a{' '}
          <StarFilledIcon className="text-yellow-500 inline h-[20px] w-[20px]" />
          on GitHub and follow us on Twitter
        </div>
        <div className="flex flex-row items-center">
          <Button
            onClick={() =>
              router.push('https://github.com/nucleuscloud/neosync')
            }
            variant="ghost"
            className="hover:bg-gray-300"
          >
            <GitHubLogoIcon />
          </Button>
          <Button
            onClick={() => router.push('https://twitter.com/neosynccloud')}
            variant="ghost"
            className="hover:bg-gray-300 "
          >
            <TwitterLogoIcon />
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
    <div className="top-0 flex md:hidden lg:hidden flex-row gap-3 justify-center w-full h-[35px] items-centerbg-gradient-to-r from-indigo-200 to-slate-300 font-satoshi">
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
