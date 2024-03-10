'use client';
import { ExternalLinkIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import { posthog } from 'posthog-js';
import { ReactElement } from 'react';
import { Badge } from '../ui/badge';

export default function ValueProps(): ReactElement {
  const features = [
    {
      title: 'Unblock local development ',
      image: '/images/unblocklocal.svg',
      link: '/solutions/unblock-local-development',
    },
    {
      title: 'Fix broken staging environments',
      image: '/images/brokenenv.svg',
      link: '/solutions/fix-staging-environments',
    },
    {
      title: 'Keep your environments in sync',
      image: '/images/envsync3.svg',
      link: '/solutions/keep-environments-in-sync',
    },
    {
      title: `Frictionless privacy and compliance`,
      image: '/images/comp.svg',
      link: '/solutions/security-privacy',
    },
  ];

  return (
    <div>
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Protect Your Sensitive Data, Build Faster with Neosync
      </div>
      <div className="text-lg text-gray-400 font-satoshi font-light pt-20 flex flex-col md:flex-row lg:flex-row gap-6 justify-center items-center">
        {features.map((item) => (
          <Link
            href={item.link}
            key={item.title}
            target="_blank"
            onClick={() => {
              posthog.capture('user click', {
                page: item.title,
              });
            }}
          >
            <div className="border border-gray-400 bg-white rounded-xl p-2 lg:p-4 xl:p-8 shadow-xl flex flex-col justify-between gap-6 text-center w-full  max-w-[300px] lg:max-w-[400px] mx-auto hover:shadow-gray-400">
              <div>
                <div className="flex justify-center">
                  <Image
                    src={item.image}
                    alt="NeosyncLogo"
                    width="250"
                    height="172"
                    className="w-[250px] h-[172px]"
                  />
                </div>
                <div className="text-gray-900 text-2xl pt-10">{item.title}</div>
              </div>
              <div className="flex justify-center">
                <Badge variant="outline">
                  Learn more <ExternalLinkIcon className="w-3 h-3 ml-2" />
                </Badge>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
