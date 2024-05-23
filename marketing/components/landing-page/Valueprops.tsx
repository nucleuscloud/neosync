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
      title: 'Safely test your code against Production data',
      image: '/images/vpt1.svg',
      link: '/solutions/unblock-local-development',
    },
    {
      title: 'Easily reproduce Production bugs locally',
      image: '/images/vp3-new.svg',
      link: '/solutions/reproduce-prod-bugs-locally',
    },
    {
      title: 'Fix broken staging environments',
      image: '/images/vp2.svg',
      link: '/solutions/fix-staging-environments',
    },
    {
      title: `Reduce your compliance scope`,
      image: '/images/comp.svg',
      link: '/solutions/security-privacy',
    },
  ];

  return (
    <div>
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center ">
        A Better Developer Experience to Ship More Resilient Code, Faster
      </div>
      <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-20 text-center">
        Safely work with anonymized Production data without any of the security
        and privacy risk for a better building and testing experience.
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
            <div className="border border-gray-300 bg-white rounded-xl p-2 lg:p-4 xl:p-8 shadow-xl flex flex-col justify-between gap-6 text-center w-full max-w-[300px] lg:w-[270px] lg:max-w-[400px] mx-auto sm:h-[100px] md:h-[360px] hover:border-gray-400">
              <div>
                <div className="flex justify-center">
                  <Image
                    src={item.image}
                    alt="NeosyncLogo"
                    width="178"
                    height="131"
                    className="w-[178px] h-[131px]"
                  />
                </div>
                <div className="text-gray-900 text-2xl pt-6">{item.title}</div>
              </div>
              <div className="flex justify-center">
                <Badge variant="outline" className="border-gray-500 ">
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
