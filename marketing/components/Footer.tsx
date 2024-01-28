import {
  DiscordLogoIcon,
  GitHubLogoIcon,
  LinkedInLogoIcon,
  TwitterLogoIcon,
} from '@radix-ui/react-icons';
import { ArrowRightIcon } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import { AiOutlineCopyright } from 'react-icons/ai';
import { Button } from './ui/button';

export default function Footer(): ReactElement {
  return (
    <div className="py-10 rounded-t-2xl bg-[#1E1E1E]">
      <div className="flex flex-col w-full justify-between gap-20 px-5 sm:px-10 md:px-20 lg:px-40  max-w-[1800px] mx-auto ">
        <div className="p-10 border border-[#484848] bg-[#282828] rounded-xl mt-10 justify-center">
          <div className="flex flex-col lg:flex-row gap-10 items-cener">
            <div className="flex flex-col gap-6 ">
              <div className="font-satoshi text-3xl lg:text-4xl text-gray-100">
                Join our Community
              </div>
              <div className="font-satoshi text-xl text-gray-400 lg:w-[80%]">
                Have questions about Neosync? Come chat with us on Discord!
              </div>
              <div className="flex flex-col sm:flex-row gap-4 lg:gap-4">
                <div>
                  <Button className="px-4" variant="secondary">
                    <Link href="https://discord.gg/UVmPTzn7dV">
                      <div className="flex flex-row items-center">
                        <p>Join our Discord</p>
                        <ArrowRightIcon className="ml-3 h-5 w-15" />
                      </div>
                    </Link>
                  </Button>
                </div>
                <div>
                  <Button className="px-8 bg-transparent border border-gray-600 hover:bg-[#303030]">
                    <Link href="https://github.com/nucleuscloud/neosync">
                      <div className="flex flex-row items-center">
                        <p>Star Neosync</p>
                        <GitHubLogoIcon className="ml-3 h-5 w-15" />
                      </div>
                    </Link>
                  </Button>
                </div>
              </div>
            </div>
            <div className="opacity-60">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/devcommunity.svg"
                alt="dev"
                width="600"
                height="191"
              />
            </div>
          </div>
        </div>
        <div className="flex flex-col lg:flex-row items-center justify-between gap-4">
          <div className="flex flex-col justify-between gap-8 items-center lg:items-start">
            <div className="flex flex-row items-center">
              <Link href="/">
                <Image
                  src="https://assets.nucleuscloud.com/neosync/newbrand/logo_text_dark_mode.svg"
                  alt="NeosyncLogo"
                  className=""
                  width="124"
                  height="80"
                />
              </Link>
            </div>
            <Socials />
          </div>
          <div>
            <Image
              src="https://assets.nucleuscloud.com/neosync/marketingsite/soc2.png"
              alt="soc2"
              className="object-scale-down"
              width="100"
              height="100"
            />
          </div>
        </div>
        <Links />
      </div>
    </div>
  );
}

function Socials(): ReactElement {
  const links = [
    {
      name: 'Github',
      logo: <GitHubLogoIcon className="w-6 h-6" />,
      href: 'https://github.com/nucleuscloud/neosync',
    },
    {
      name: 'X',
      logo: <TwitterLogoIcon className="w-6 h-6" />,
      href: 'https://twitter.com/neosynccloud',
    },
    {
      name: 'Linkedin',
      logo: <LinkedInLogoIcon className="w-6 h-6" />,
      href: 'https://linkedin.com/company/neosync',
    },
    {
      name: 'Discord',
      logo: <DiscordLogoIcon className="w-6 h-6" />,
      href: 'https://discord.gg/UVmPTzn7dV',
    },
  ];
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row gap-2 items-center">
        {links.map((item) => (
          <Button key={item.name} variant="footerlink">
            <Link href={item.href}>{item.logo}</Link>
          </Button>
        ))}
      </div>
    </div>
  );
}

function Links(): ReactElement {
  return (
    <div className="flex flex-col w-full lg:pt-10 ">
      <div className="h-[2px] rounded-full w-full bg-[#303030]" />
      <div className="flex flex-col lg:flex-row lg:justify-center items-endspace-y-3 lg:space-x-8 pt-10">
        <div className="flex flex-row items-center">
          <AiOutlineCopyright className="text-gray-500" />
          <div className="text-gray-500 font-satoshi text-sm pl-2">
            Nucleus Cloud Corp. {new Date().getFullYear()}
          </div>
        </div>
        <Link className="pl-0 font-satoshi no-underline" href="/privacy-policy">
          <div className="text-gray-500 text-sm hover:text-gray-200">
            Privacy Policy
          </div>
        </Link>
        <Link
          className="pl-0 font-satoshi no-underline"
          href="/terms-of-service"
        >
          <div className="text-gray-500 text-sm hover:text-gray-200">
            Terms of Service
          </div>
        </Link>
      </div>
    </div>
  );
}
