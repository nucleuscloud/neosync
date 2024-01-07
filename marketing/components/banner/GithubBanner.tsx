'use client';
import {
  GitHubLogoIcon,
  StarFilledIcon,
  TwitterLogoIcon,
} from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { Button } from '../ui/button';

export default function GithubBanner() {
  const router = useRouter();
  return (
    <div className=" top-0 flex flex-row gap-3 justify-center w-full h-[35px]-400 items-center bg-[#e5edf6]">
      <div>
        If you like Neosync, give it a{' '}
        <StarFilledIcon className="text-yellow-500 inline h-[20px] w-[20px]" />{' '}
        on GitHub and follow us on Twitter
      </div>
      <div className="flex flex-row gap-2 items-center">
        <Button
          onClick={() => router.push('https://github.com/nucleuscloud/neosync')}
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
  );
}
