'use client';
import Image from 'next/image';
import Link from 'next/link';
import { posthog } from 'posthog-js';
import { ReactElement } from 'react';
import { Badge } from '../ui/badge';

export default function ValueProps(): ReactElement {
  const features = [
    {
      title: 'Unblock local development ',
      description:
        'Self-serve de-identified and synthetic data locally without worrying about sensitive data privacy or security. ',
      image: '/images/unblocklocal.svg',
      link: '/solutions/unblock-local-development',
    },
    {
      title: 'Fix broken staging environments',
      description:
        'Catch bugs before production when you hydrate your staging and QA environments with production-like data. ',
      image: '/images/brokenenv.svg',
      link: '/solutions/fix-staging-environments',
    },
    {
      title: 'Keep your environments in sync',
      description:
        'Speed up your dev and test cycles by keeping your environments in sync with the latest de-identified data set.',
      image: '/images/envsync3.svg',
      link: '/solutions/keep-environments-in-sync',
    },
    {
      title: `Frictionless privacy and compliance`,
      description: `Comply with HIPAA, GDPR, and DPDP with de-identified and synthetic data.`,
      image: '/images/comp.svg',
      link: '/solutions/security-privacy',
    },
  ];

  return (
    <div>
      <div className="px-6">
        <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
          Protect Your Sensitive Data, Build Faster with Neosync
        </div>
      </div>
      <div className="text-lg text-gray-400 font-satoshi font-light  pt-10 lg:pt-20 flex flex-col md:flex-row lg:flex-row gap-6 justify-center items-center">
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
            <div className="border border-gray-400 bg-white rounded-xl p-2 lg:p-4 xl:p-8 shadow-xl flex flex-col justify-between gap-6 text-center w-full  max-w-[300px] lg:max-w-[400px] mx-auto md:h-[560px] lg:h-[560px] hover:shadow-gray-400">
              <div>
                <div className="flex justify-center">
                  <Image
                    src={item.image}
                    alt="NeosyncLogo"
                    width="250"
                    height="172"
                  />
                </div>
                <div className="text-gray-900 text-2xl pt-10">{item.title}</div>
                <div className=" text-gray-500 text-[16px] pt-6 text-center">
                  {item.description}
                </div>
              </div>
              <div className="flex justify-center">
                <Badge variant="outline">Learn more</Badge>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
